package main

import (
	"fmt"

	"github.com/nickwells/pusu.mod/pusu"
)

// subsMap is the type representing a map between a topic and a collection of
// clients who are subscribed to that topic. There is one such map per
// namespace
type subsMap map[pusu.Topic]map[*client]bool

// namespaceSubsMap is the type representing a map between a topic and the
// collection of clients who are subscribed to that topic
type namespaceSubsMap map[pusu.Namespace]subsMap

// serverMsgHandler is a function for handling a message from a server
// perspective
type serverMsgHandler func(*prog, clientMessage, namespaceSubsMap)

// serverMsgHandlerMap maps message types to per-type handlers
type serverMsgHandlerMap map[pusu.MsgType]serverMsgHandler

// serverProtocolError generates a server message handler function which logs
// a message stating that there has been a protocol error.
func serverProtocolError(problem string) serverMsgHandler {
	return func(prog *prog, cMsg clientMessage, _ namespaceSubsMap) {
		err := fmt.Errorf("message type %q is not allowed - %s",
			cMsg.msg.MT, problem)
		prog.logger.Error("protocol error", pusu.ErrorAttr(err))
	}
}

// serverHandlerBadMT reports a protocol error - an unknown message type
func serverHandlerBadMT(prog *prog, cMsg clientMessage, _ namespaceSubsMap) {
	err := fmt.Errorf("unknown message type %q", cMsg.msg.MT)
	prog.logger.Error("protocol error", pusu.ErrorAttr(err))
}

// setAllEntries sets all the entries to the suplied handler
func (hm serverMsgHandlerMap) setAllEntries(mh serverMsgHandler) {
	for mt := range pusu.MaxMsgType {
		if err := mt.Check(); err != nil {
			continue
		}

		hm[mt] = mh
	}
}

// setEntries sets the specified handlerMap entries to the suplied handler
func (hm serverMsgHandlerMap) setEntries(
	mh serverMsgHandler,
	mts ...pusu.MsgType,
) {
	for _, mt := range mts {
		hm[mt] = mh
	}
}

// getHandler returns a message handler for the given message type
func (hm serverMsgHandlerMap) getHandler(mt pusu.MsgType) serverMsgHandler {
	var handler serverMsgHandler

	var ok bool

	if handler, ok = hm[mt]; !ok {
		handler = serverHandlerBadMT
	}

	return handler
}
