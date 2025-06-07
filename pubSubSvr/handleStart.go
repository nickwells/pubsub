package main

import (
	"fmt"

	"github.com/nickwells/pusu.mod/pusu"
)

// setNamespace sets the client namespace from the passed string. It returns
// a non-nil error if the server does not allow the supplied namespace.
func (clt *client) setNamespace(n string) error {
	clt.namespace = pusu.Namespace(n)
	if !clt.nsRules.isValid(clt.namespace) {
		err := fmt.Errorf("bad namespace: %s", clt.namespace)

		clt.logger.Error("namespace not allowed by this server",
			clt.namespace.Attr(),
			pusu.ErrorAttr(err))

		return err
	}

	return nil
}

// setProtoVsn sets the client protocol version from the passed value. It returns
// a non-nil error if the server does not allow the supplied protocol version.
func (clt *client) setProtoVsn(pv int32) error {
	var err error

	clt.protoVsn = pusu.ProtoVsn(pv)

	switch {
	case clt.protoVsn > pusu.CurrentProtoVsn:
		err = fmt.Errorf(
			"bad client protocol version; maximum supported version: %d",
			pusu.CurrentProtoVsn)
	case clt.protoVsn < pusu.CurrentProtoVsn:
		err = fmt.Errorf(
			"bad client protocol version; minimum supported version: %d",
			pusu.CurrentProtoVsn)
	}

	if err != nil {
		clt.logger.Error("protocol version not allowed by this server",
			clt.protoVsn.Attr(),
			pusu.ErrorAttr(err))
	}

	return err
}

// clientHandleStart handles the Start message (the first message that the
// client should send).
func clientHandleStart(clt *client, msg *pusu.Message) error {
	clt.logger.Info("client handling message", msg.MT.Attr(), msg.MsgID.Attr())

	smp := pusu.StartMsgPayload{}
	if err := msg.Unmarshal(&smp, clt.logger); err != nil {
		return err
	}

	clt.identity = smp.ClientId

	if err := clt.setNamespace(smp.Namespace); err != nil {
		return err
	}

	if err := clt.setProtoVsn(smp.ProtocolVersion); err != nil {
		return err
	}

	clt.logger.Info("client start information",
		clt.startInfoAttr(), clt.namespace.Attr(), clt.protoVsn.Attr())

	// disable any future messages of this type ...
	clt.handlers.setAllEntries(
		clientProtocolError("the client should not send this type of message"))
	clt.handlers.setEntries(
		clientProtocolError("only the first message may be of this type"),
		pusu.Start)

	// ... and set the remaining message handlers
	clt.handlers.setEntries(clientHandlePublish, pusu.Publish)
	clt.handlers.setEntries(clientHandleSubscribe, pusu.Subscribe)
	clt.handlers.setEntries(clientHandleUnsubscribe, pusu.Unsubscribe)
	clt.handlers.setEntries(clientHandlePing, pusu.Ping)

	clt.sendAck(msg.MsgID)

	return nil
}
