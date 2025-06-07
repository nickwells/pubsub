package main

import (
	"github.com/nickwells/pusu.mod/pusu"
)

// clientHandlePublish handles the publish message from the client side. It
// simply hands it on to the server over the pubSubChan.
func clientHandlePublish(clt *client, msg *pusu.Message) error {
	clt.logger.Info("client handling message", msg.MT.Attr(), msg.MsgID.Attr())

	clt.pubSubChan <- clientMessage{
		clt: clt,
		msg: msg,
	}

	clt.sendAck(msg.MsgID)

	return nil
}

// serverHandlePublish handles a Publish message from the server side
//
// Note that this handler takes the clientMessage sent over the pubSubChan by
// the clientHandlePublish func.
func serverHandlePublish(
	prog *prog,
	cMsg clientMessage,
	nsm namespaceSubsMap,
) {
	prog.logger.Info("server handling message",
		cMsg.msg.MT.Attr(), cMsg.msg.MsgID.Attr())

	pmp := pusu.PublishMsgPayload{}
	if err := cMsg.msg.Unmarshal(&pmp, prog.logger); err != nil {
		cMsg.clt.sendError(cMsg.msg.MsgID, err)

		return
	}

	var topicSubs subsMap

	var cMap map[*client]bool

	var ok bool

	if topicSubs, ok = nsm[cMsg.clt.namespace]; !ok {
		return
	}

	subTopics := pusu.Topic(pmp.Topic).SubTopics()

SubTopicLoop:
	for _, topic := range subTopics {
		if cMap, ok = topicSubs[topic]; !ok {
			continue SubTopicLoop // no subscriptions
		}

		msg := pusu.Message{
			MT: pusu.Publish,
		}

		pmp.Topic = string(topic)

		if err := (&msg).Marshal(&pmp, prog.logger); err != nil {
			return
		}

		for clt := range cMap {
			clt.sendMessage(msg)
		}
	}
}
