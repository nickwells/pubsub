package main

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/nickwells/pusu.mod/pusu"
	"github.com/nickwells/verbose.mod/verbose"
)

// clientMessage holds a message and the client who sent it
type clientMessage struct {
	clt *client
	msg *pusu.Message
}

// prog holds program parameters and status
type prog struct {
	exitStatus int
	stack      *verbose.Stack

	// parameters
	port                    int           // the port number to listen on
	logDir                  string        // the directory for the log files
	statusReportingInterval time.Duration // how long between status reports
	certInfo                pusu.CertInfo // certificates
	logLevel                slog.Level    // level at which to log messages
	progName                string        // the name of the program

	nsRules namespaceRules // the rules governing which namespaces are valid

	// program data
	logger *slog.Logger

	handlers serverMsgHandlerMap

	listener  net.Listener
	tlsConfig *tls.Config

	pubSubChan     chan clientMessage
	disconnectChan chan *client
}

// newProg returns a new Prog instance with the default values set
func newProg() *prog {
	const dfltStatusInterval = 5

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("cannot get the user home directory: %w", err))
	}

	return &prog{
		stack: &verbose.Stack{},

		logDir:                  filepath.Join(homeDir, "logs"),
		statusReportingInterval: dfltStatusInterval * time.Second,
		logLevel:                slog.LevelInfo,
		handlers:                make(serverMsgHandlerMap),
		pubSubChan:              make(chan clientMessage),
		disconnectChan:          make(chan *client),
	}
}

// setExitStatus sets the exit status to the new value. It will not do this
// if the exit status has already been set to a non-zero value.
func (prog *prog) setExitStatus(es int) {
	if prog.exitStatus == 0 {
		prog.exitStatus = es
	}
}

// startLogger initialises the logger for the program
func (prog *prog) startLogger() {
	prog.logger = slog.New(
		slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{Level: prog.logLevel},
		))
}

// openListener constructs the tls listener. Any errors will be logged, will
// set the exitStatus to non-zero and this will return false. If all the
// steps succeed this will return true.
func (prog *prog) openListener() bool {
	if err := prog.certInfo.PopulateCert(); err != nil {
		prog.logger.Error("couldn't populate the server certificate",
			pusu.ErrorAttr(err))
		prog.setExitStatus(1)

		return false
	}

	if err := prog.certInfo.PopulateCertPool(); err != nil {
		prog.logger.Error("couldn't populate the server's certificate pool",
			pusu.ErrorAttr(err))
		prog.setExitStatus(1)

		return false
	}

	prog.tlsConfig = &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    prog.certInfo.CertPool(),
		Certificates: []tls.Certificate{prog.certInfo.Cert()},
		MinVersion:   tls.VersionTLS13,
	}

	var err error

	laddr := net.JoinHostPort("localhost", fmt.Sprintf("%d", prog.port))

	prog.listener, err = tls.Listen("tcp", laddr, prog.tlsConfig)
	if err != nil {
		prog.logger.Error("couldn't make the tls Listener",
			pusu.ErrorAttr(err))
		prog.setExitStatus(1)

		return false
	}

	prog.logger.Info("listening", listeningPortAttr(prog.port))

	return true
}

// reportAllowedNamespaces prints log messages listing the allowed namespace
// values
func (prog *prog) reportAllowedNamespaces() {
	if len(prog.nsRules.allowed) > 0 {
		prog.logger.Info("limited namespaces - only those matching are allowed",
			slog.Int("namespace-count", len(prog.nsRules.allowed)))

		for n := range prog.nsRules.allowed {
			prog.logger.Info("allowed namespace", n.Attr())
		}

		return
	}

	if len(prog.nsRules.prefixes) > 0 {
		prog.logger.Info(
			"limited namespaces - it must have one of the following prefixes",
			slog.Int("prefix-count", len(prog.nsRules.prefixes)))

		for _, pfx := range prog.nsRules.prefixes {
			prog.logger.Info("allowed prefixes", slog.String("prefix", pfx))
		}

		return
	}

	prog.logger.Info("any namespace is allowed")
}

// run is the starting point for the program, it should be called from main()
// after the command-line parameters have been parsed. Use the setExitStatus
// method to record the exit status and then main can exit with that status.
func (prog *prog) run() {
	prog.startLogger()

	prog.logger.Info("starting", progNameAttr(prog.progName))
	prog.reportAllowedNamespaces()

	if !prog.openListener() {
		return
	}

	defer func() {
		prog.logger.Info("closing the listener", listeningPortAttr(prog.port))

		if err := prog.listener.Close(); err != nil {
			prog.logger.Error("problem closing the listener",
				pusu.ErrorAttr(err))
		} else {
			prog.logger.Info("listener closed")
		}
	}()

	var cID connID

	go prog.pubSubHandler()

	for {
		conn, err := prog.listener.Accept()
		if err != nil {
			prog.logger.Error("couldn't Accept the client connection",
				pusu.ErrorAttr(err))

			continue
		}

		cID++

		startClient(prog.logger,
			cID,
			conn,
			prog.pubSubChan,
			prog.disconnectChan,
			prog.nsRules)
	}
}

// setAllHandlers populates the server-side message handlers
func (prog *prog) setAllHandlers() {
	prog.handlers.setAllEntries(serverProtocolError(
		"only publish, subscribe or unsubscribe messages are expected"))
	prog.handlers.setEntries(serverHandlePublish, pusu.Publish)
	prog.handlers.setEntries(serverHandleSubscribe, pusu.Subscribe)
	prog.handlers.setEntries(serverHandleUnsubscribe, pusu.Unsubscribe)
}

// pubSubHandler handles all the pings, publications, subscriptions and
// unsubscribes
func (prog *prog) pubSubHandler() {
	subscriptions := make(namespaceSubsMap)
	ticker := time.NewTicker(prog.statusReportingInterval)
	msgTypeCount := map[pusu.MsgType]int{}

	prog.setAllHandlers()

	for {
		select {
		case cMsg := <-prog.pubSubChan:
			prog.logger.Info("server message received",
				cMsg.clt.cID.Attr(), cMsg.msg.MT.Attr(), cMsg.msg.MsgID.Attr())

			msgTypeCount[cMsg.msg.MT]++

			handler := prog.handlers.getHandler(cMsg.msg.MT)

			handler(prog, cMsg, subscriptions)

		case clt := <-prog.disconnectChan:
			prog.logger.Info("server client disconnection received",
				clt.cID.Attr())
			removeClientSubs(clt, subscriptions)

		case <-ticker.C:
			go logStatus(prog, msgTypeCount, len(subscriptions))

			msgTypeCount = map[pusu.MsgType]int{}
		}
	}
}

// removeClientSubs removes all the subscriptions that the client has
func removeClientSubs(clt *client, subscriptions namespaceSubsMap) {
	if len(clt.subs) != 0 {
		subsMap := subscriptions[clt.namespace]
		for t := range clt.subs {
			cs := subsMap[t]
			delete(cs, clt)

			if len(cs) == 0 {
				delete(subsMap, t)

				if len(subsMap) == 0 {
					delete(subscriptions, clt.namespace)
				}
			}
		}
	}
}

// logStatus reports the status of the server
func logStatus(prog *prog, msgTypeCount map[pusu.MsgType]int, subsCount int) {
	attrs := make([]any, 0, pusu.MaxMsgType-1)

	for mt := range pusu.MaxMsgType {
		if err := mt.Check(); err != nil {
			continue
		}

		attrs = append(attrs,
			slog.Int((mt).String(), msgTypeCount[mt]))
	}

	counts := slog.Group("msgType", attrs...)

	prog.logger.Info("status", counts)
	prog.logger.Info("subscriptions", slog.Int("namespaces", subsCount))
}
