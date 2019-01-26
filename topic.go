package twitchpubsub

import "fmt"

type topicHash string

type websocketTopic struct {
	name      string
	connected bool
	authToken string
	hash      topicHash

	// Nonce used when establishing a connection to this topic
	// If a topic has a nonce, it implies that it is currently owned by a connection
	nonce string
}

func hashTopic(t *websocketTopic) topicHash {
	return topicHash(fmt.Sprintf("%s:%s", t.name, t.authToken))
}

func newTopic(name, authToken string) *websocketTopic {
	t := &websocketTopic{
		name:      name,
		authToken: authToken,
	}
	t.hash = hashTopic(t)
	return t
}
