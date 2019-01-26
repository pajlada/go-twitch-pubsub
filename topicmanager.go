package twitchpubsub

import "sync"

type topicManager struct {
	mutex  *sync.Mutex
	topics map[topicHash]*websocketTopic
}

func newTopicManager() *topicManager {
	return &topicManager{
		mutex:  &sync.Mutex{},
		topics: make(map[topicHash]*websocketTopic),
	}
}

func (t *topicManager) Add(topic *websocketTopic) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.topics[topic.hash]; ok {
		// We are already subscribed to this topic
		return false
	}
	t.topics[topic.hash] = topic
	return true
}
