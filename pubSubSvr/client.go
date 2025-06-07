package main

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/nickwells/pusu.mod/pusu"
)

// client represents a client of the server - a connection from another
// program. The identity is supplied by the connecting client and is not
// verified or validated; it should not be trusted
type client struct {
	sync.Mutex

	// cID is the unique id number for the client connection
	cID connID
	// identity is the information string passed from the client to the
	// server
	identity string
	// namespace is the namespace in which all of the client subscription
	// topics lie
	namespace pusu.Namespace
	// protoVsn is the version of the communication protocol that
	// this client will use
	protoVsn pusu.ProtoVsn

	logger *slog.Logger

	conn      net.Conn
	connected bool

	subs     map[pusu.Topic]bool
	handlers clientMsgHandlerMap

	pubSubChan     chan clientMessage
	disconnectChan chan *client

	sendChan chan pusu.Message

	nsRules namespaceRules
}

// startClient returns a pointer to a newly instantiated client
func startClient(
	logger *slog.Logger,
	cid connID,
	conn net.Conn,
	psChan chan clientMessage,
	disconnectChan chan *client,
	nsRules namespaceRules,
) {
	const maxBacklog = 20

	clt := &client{
		cID:            cid,
		conn:           conn,
		subs:           make(map[pusu.Topic]bool),
		handlers:       make(clientMsgHandlerMap),
		pubSubChan:     psChan,
		disconnectChan: disconnectChan,
		sendChan:       make(chan pusu.Message, maxBacklog),
		nsRules:        nsRules,
	}

	clt.logger = logger.With(cid.Attr())

	clt.handlers.setAllEntries(
		clientProtocolError("the first message must be of type " +
			pusu.Start.String()))
	clt.handlers.setEntries(clientHandleStart, pusu.Start)

	clt.logger.Info("connection received", netAddrAttr(clt.conn))

	var wg sync.WaitGroup

	wg.Add((2))

	go clt.reader(&wg)
	go clt.writer(&wg)

	wg.Wait()

	clt.connected = true
}

// startInfoAttr returns a standardised slog Attr giving the client identity
// details as received in the Start message.
func (clt *client) startInfoAttr() slog.Attr {
	return slog.String(cltAttrPfx+"Start-Info", clt.identity)
}

// readMsg reads the next message received over the client's connection
func (clt *client) readMsg() (pusu.Message, error) {
	return pusu.ReadMsg(clt.conn)
}

// reader reads from the connection repeatedly and handles the messages
// received.
func (clt *client) reader(wg *sync.WaitGroup) {
	clt.logger.Info("reader started")

	wg.Done()

Loop:
	for {
		msg, err := clt.readMsg()
		if err != nil {
			clt.handleReadError(err)

			break Loop
		}

		clt.logger.Info("client message received", msg.MT.Attr())

		handler, ok := clt.handlers[msg.MT]

		if !ok {
			err := errors.New("unexpected message type: " + msg.MT.String())

			clt.logger.Error("couldn't handle client message",
				pusu.ErrorAttr(err))
			clt.sendError(msg.MsgID, err)

			break Loop
		}

		if err = handler(clt, &msg); err != nil {
			clt.logger.Error("the client message handler failed",
				pusu.ErrorAttr(err),
				msg.MT.Attr())
			clt.sendError(msg.MsgID, err)

			break Loop
		}
	}

	clt.logger.Info("reader finished")
}

// writer listens on a channel and writes the messages to the client
func (clt *client) writer(wg *sync.WaitGroup) {
	clt.logger.Info("writer started")

	wg.Done()

Loop:
	for msg := range clt.sendChan {
		if err := msg.Write(clt.conn); err != nil {
			clt.logger.Error("couldn't write the message to the client",
				msg.MT.Attr(),
				pusu.ErrorAttr(err))

			break Loop
		}

		if msg.MT == pusu.Error {
			clt.disconnect()
		}
	}

	clt.logger.Info("writer finished")
}

// handleReadError reports an error detected when reading from the client
// connection. It then notifies the server that the client is disconnecting.
func (clt *client) handleReadError(err error) {
	if err == nil {
		return
	}

	if errors.Is(err, io.EOF) {
		clt.logger.Info("client disconnected")
	} else {
		clt.logger.Error("could not read the client message",
			pusu.ErrorAttr(err))
	}

	clt.disconnectChan <- clt
}

// closeConn closes the client connection; any errors detected
// while doing so are ignored.
func (clt *client) closeConn() {
	clt.logger.Info("closing connection")

	_ = clt.conn.Close()

	clt.logger.Info("connection closed")
}

// sendError sends the error to the client as an error message
func (clt *client) sendError(msgID pusu.MsgID, err error) {
	if err == nil {
		err = errors.New("unknown error")
	}

	msg := pusu.Message{
		MT:    pusu.Error,
		MsgID: msgID,
	}

	_ = (&msg).Marshal(&pusu.ErrorMsgPayload{
		Error: err.Error(),
	}, clt.logger)

	clt.sendMessage(msg)
}

// sendAck sends an ack message to the client
func (clt *client) sendAck(msgID pusu.MsgID) {
	clt.sendMessage(pusu.Message{
		MT:    pusu.Ack,
		MsgID: msgID,
	})
}

// sendMessage checks that the sendChan is not full and closes it if it
// is. Otherwise it writes the message to the channel.
func (clt *client) sendMessage(msg pusu.Message) {
	clt.Lock()
	defer clt.Unlock()

	if !clt.connected {
		return
	}

	select {
	case clt.sendChan <- msg:
	default:
		clt.logger.Error("slow consumer")

		clt.closeConn()
		close(clt.sendChan)
		clt.connected = false
	}
}

// disconnect handles the disconnection behaviour
func (clt *client) disconnect() {
	clt.Lock()
	defer clt.Unlock()

	if !clt.connected {
		return
	}

	clt.closeConn()
	close(clt.sendChan)
	clt.connected = false
}
