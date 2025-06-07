package main

import "github.com/nickwells/pusu.mod/pusu"

// clientHandlePing handles the ping message. It simply hands it on to the
// server over the pubSubChan.
func clientHandlePing(clt *client, msg *pusu.Message) error {
	clt.logger.Info("client handling message", msg.MT.Attr(), msg.MsgID.Attr())

	clt.sendMessage(*msg)

	return nil
}
