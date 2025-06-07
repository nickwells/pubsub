package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/pusu.mod/pusu"
	"github.com/nickwells/pusu.mod/pusuclt"
	"github.com/nickwells/verbose.mod/verbose"
)

// prog holds program parameters and status
type prog struct {
	exitStatus int
	stack      *verbose.Stack

	// parameters
	cci       *pusuclt.ConnInfo
	logLevel  slog.Level
	namespace pusu.Namespace
	topic     pusu.Topic

	payload      string
	payloadParam *param.ByName

	timeout      time.Duration
	timeoutParam *param.ByName

	count      int
	countParam *param.ByName

	progName string

	// program data
	conn *pusuclt.Client

	logger *slog.Logger
}

// newProg returns a new Prog instance with the default values set
func newProg() *prog {
	const defaultTimeOutSecs = 2

	return &prog{
		stack:   &verbose.Stack{},
		cci:     pusuclt.NewConnInfo(nil),
		count:   1,
		timeout: time.Duration(defaultTimeOutSecs * time.Second),
	}
}

// startLogger initialises the logger for the program
func (prog *prog) startLogger() {
	prog.logger = slog.New(
		slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{Level: prog.logLevel},
		)).With(slog.Int("pid", os.Getpid()))
}

// setExitStatus sets the exit status to the new value. It will not do this
// if the exit status has already been set to a non-zero value.
func (prog *prog) setExitStatus(es int) {
	if prog.exitStatus == 0 {
		prog.exitStatus = es
	}
}

// run is the starting point for the program, it should be called from main()
// after the command-line parameters have been parsed. Use the setExitStatus
// method to record the exit status and then main can exit with that status.
func (prog *prog) run() {
	var err error

	prog.startLogger()

	if prog.conn, err = pusuclt.NewClient(
		prog.namespace, prog.progName,
		prog.logger, prog.cci); err != nil {
		prog.logger.Error("could not construct the client connection",
			pusu.ErrorAttr(err))
		prog.setExitStatus(1)

		return
	}

	if prog.payloadParam.HasBeenSet() {
		prog.sendAll()
	} else {
		prog.recvAll()
	}

	prog.logger.Info("finished")
}

// sendAll publishes the messages it returns when all the messages are sent.
func (prog *prog) sendAll() {
	errSendTimeout := errors.New("time-out before all messages were sent")

	var sendCount int

	msgPubChan := make(chan int)

	cb := pusuclt.MakeCallback(msgPubChan, 1)

	go prog.publish(cb)

PublishLoop:
	for {
		select {
		case <-msgPubChan:
			sendCount++
			if sendCount == prog.count {
				prog.logger.Info("all messages published",
					slog.Int("message-count", sendCount))
				break PublishLoop
			}
		case <-time.After(prog.timeout):
			prog.logger.Error("timedout", pusu.ErrorAttr(errSendTimeout))
			break PublishLoop
		}
	}
}

// recvAll subscribes and waits for all the messages to be received
func (prog *prog) recvAll() {
	prog.logger.Info("subscribing")

	errRecvTimeout := errors.New("time-out before all messages were received")

	msgRecdChan := make(chan struct{})

	var mh pusuclt.MsgHandler = func(t pusu.Topic, payload []byte) {
		prog.logger.Info("received message",
			t.Attr(), slog.String("payload", string(payload)))
		msgRecdChan <- struct{}{}
	}

	err := prog.conn.Subscribe(nil, pusuclt.TopicHandler{
		Topic:   pusu.Topic(prog.topic),
		Handler: mh,
	})
	if err != nil {
		prog.logger.Error("Subscribe failed", pusu.ErrorAttr(err))
		prog.setExitStatus(1)

		return
	}

	prog.logger.Info("subscribed, waiting to receive messages")

	var receiveCount int
SubscribeLoop:
	for {
		select {
		case <-msgRecdChan:
			receiveCount++
			if receiveCount >= prog.count {
				prog.logger.Info("all messages received",
					slog.Int("message-count", receiveCount))
				break SubscribeLoop
			}
		case <-time.After(prog.timeout):
			prog.logger.Error("timedout", pusu.ErrorAttr(errRecvTimeout))
			break SubscribeLoop
		}
	}
}

// publish will publish the messages
func (prog *prog) publish(cb pusuclt.Callback) {
PublishLoop:
	for i := range prog.count {
		p := fmt.Sprintf("%s%02d", prog.payload, i)

		prog.logger.Info("Publishing",
			slog.String("payload", p),
			slog.Int("message-number", i))

		if err := prog.conn.Publish(cb, prog.topic, []byte(p)); err != nil {
			prog.logger.Error("couldn't publish the payload",
				prog.topic.Attr(), pusu.ErrorAttr(err))
			prog.setExitStatus(1)
			break PublishLoop
		}
	}
}
