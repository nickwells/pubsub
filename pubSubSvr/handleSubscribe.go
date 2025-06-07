package main

import (
	"github.com/nickwells/pusu.mod/pusu"
)

// clientHandleSubscribe handles the subscribe message. It first sends the
// Subscribe message to the pubSubChan. Then it opens the messasge and adds
// each topic to the clients own subscription map.
func clientHandleSubscribe(clt *client, msg *pusu.Message) error {
	clt.logger.Info("client handling message", msg.MT.Attr(), msg.MsgID.Attr())

	smp := pusu.SubscriptionMsgPayload{}
	if err := msg.Unmarshal(&smp, clt.logger); err != nil {
		return err
	}

	for _, sub := range smp.Subs {
		topic := pusu.Topic(sub.Topic)

		clt.subs[topic] = true
	}

	clt.pubSubChan <- clientMessage{
		clt: clt,
		msg: msg,
	}

	clt.sendAck(msg.MsgID)

	return nil
}

// serverHandleSubscribe handles a Subscribe message from the server side
//
// Note that this handler takes the clientMessage sent over the pubSubChan by
// the clientHandleSubscribe func.
func serverHandleSubscribe(
	prog *prog,
	cMsg clientMessage,
	nsm namespaceSubsMap,
) {
	prog.logger.Info("server handling message",
		cMsg.msg.MT.Attr(), cMsg.msg.MsgID.Attr())

	smp := pusu.SubscriptionMsgPayload{}
	if err := cMsg.msg.Unmarshal(&smp, prog.logger); err != nil {
		return
	}

	var topicSubs subsMap

	var cMap map[*client]bool

	var ok bool

	if topicSubs, ok = nsm[cMsg.clt.namespace]; !ok {
		topicSubs = make(subsMap)
		nsm[cMsg.clt.namespace] = topicSubs
	}

	for _, sub := range smp.Subs {
		topic := pusu.Topic(sub.Topic)

		if cMap, ok = topicSubs[topic]; !ok {
			cMap = make(map[*client]bool)
			topicSubs[topic] = cMap
		}

		cMap[cMsg.clt] = true
	}
}
