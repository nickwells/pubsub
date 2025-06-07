package main

import (
	"fmt"

	"github.com/nickwells/pusu.mod/pusu"
)

// clientMsgHandler is a function for handling a message from a client
// perspective
type clientMsgHandler func(*client, *pusu.Message) error

// clientMsgHandlerMap maps message types to per-type handlers
type clientMsgHandlerMap map[pusu.MsgType]clientMsgHandler

// clientProtocolError generates a message handler function which logs a message
// stating that there has been a protocol error and returns an error
// describing the problem
func clientProtocolError(problem string) clientMsgHandler {
	return func(clt *client, msg *pusu.Message) error {
		err := fmt.Errorf("message type %q is not allowed - %s",
			msg.MT, problem)
		clt.logger.Error("protocol error", pusu.ErrorAttr(err))

		return err
	}
}

// setAllEntries sets all the entries to the suplied handler
func (hm clientMsgHandlerMap) setAllEntries(mh clientMsgHandler) {
	for mt := range pusu.MaxMsgType {
		if err := mt.Check(); err != nil {
			continue
		}

		hm[mt] = mh
	}
}

// setEntries sets the specified handlerMap entries to the suplied handler
func (hm clientMsgHandlerMap) setEntries(
	mh clientMsgHandler,
	mts ...pusu.MsgType,
) {
	for _, mt := range mts {
		hm[mt] = mh
	}
}
