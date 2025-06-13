<!-- Created by mkdoc DO NOT EDIT. -->

# Notes

## exampleClient \- purpose
this serves as an example of how to write a client program that uses the
publish/subscribe server\. It demonstrates both Publish and Subscribe
functionality\.


## pub/sub \- Namespaces
A publish/subscribe namespace provides a way to partition topics in the message
broker \(the pub/sub server\)\.



Subscriptions and publications by clients sharing the same namespace will relate
only to those clients\. Two clients having different namespaces can each
subscribe to the same topic but they will only receive messages published on
that topic from other clients sharing their namespace\.



For instance, two clients, c1, in namespace &apos;A&apos; and c2 in namespace
&apos;B&apos; can each subscribe to topic &apos;/T&apos;\. If, subseqently, a
third client, also in namespace &apos;A&apos; publishes a message on topic
&apos;/T&apos; then only client c1 will receive the message\.


## pub/sub \- Topics
A publish/subscribe topic provides routing information for the message broker
\(the pub/sub server\) to decide where to send messages it has received\.



A client can subscribe to a topic and then any messages which are subsequently
published on that topic will be sent to that client\. To stop receiving such
messages the client will need to unsubsubscribe from that topic\.



A topic name must be a well\-formed path, starting with a &apos;/&apos; and
having one or more parts following this, each part separated from its
predecessor by a single &apos;/&apos;\.


