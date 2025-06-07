package main

import (
	"github.com/nickwells/pusu.mod/pusu"
)

// clientHandleUnsubscribe handles the unsubscribe message from the client
// side. It first hands it on to the server over the pubSubChan and then
// removes all the topics from the set of client subscriptions.
func clientHandleUnsubscribe(clt *client, msg *pusu.Message,
) error {
	clt.logger.Info("client handling message", msg.MT.Attr(), msg.MsgID.Attr())

	smp := pusu.SubscriptionMsgPayload{}
	if err := msg.Unmarshal(&smp, clt.logger); err != nil {
		return err
	}

	for _, sub := range smp.Subs {
		topic := pusu.Topic(sub.Topic)

		delete(clt.subs, topic)
	}

	clt.pubSubChan <- clientMessage{
		clt: clt,
		msg: msg,
	}

	clt.sendAck(msg.MsgID)

	return nil
}

// serverHandleUnsubscribe handles an Unsubscribe message from the server
// side. It removes the client from the set of clients subscribed to the
// topic and if that leaves the set of clients subscribed to a topic empty
// then it removes the topic from the set of topic subscriptions as well.
//
// Note that this handler takes the clientMessage sent over the pubSubChan by
// the clientHandleUnsubscribe func.
func serverHandleUnsubscribe(
	prog *prog,
	cMsg clientMessage,
	nsm namespaceSubsMap,
) {
	prog.logger.Info("server handling message",
		cMsg.msg.MT.Attr(), cMsg.msg.MsgID.Attr())

	smp := pusu.SubscriptionMsgPayload{}

	if err := cMsg.msg.Unmarshal(&smp, prog.logger); err != nil {
		cMsg.clt.sendError(cMsg.msg.MsgID, err)

		return
	}

	var topicSubs subsMap

	var cMap map[*client]bool

	var ok bool

	if topicSubs, ok = nsm[cMsg.clt.namespace]; !ok {
		return
	}

	for _, sub := range smp.Subs {
		topic := pusu.Topic(sub.Topic)

		if cMap, ok = topicSubs[topic]; ok {
			delete(cMap, cMsg.clt)

			if len(cMap) == 0 {
				delete(topicSubs, topic)
			}
		}
	}
}
