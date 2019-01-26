package twitchpubsub

import (
	"fmt"
	"sync"
)

type connectionManager struct {
	host string

	// TODO: mutex lock connections slice
	connections []*connection

	// Max number of active connections
	connectionLimit      int
	connectionLimitMutex *sync.RWMutex

	// Max number of topics per connection
	topicLimit      int
	topicLimitMutex *sync.RWMutex

	messageBus  messageBusType
	quitChannel chan struct{}
}

func newConnectionManager(host string, messageBus messageBusType, quitChannel chan struct{}) *connectionManager {
	return &connectionManager{
		host: host,

		connectionLimit: 10,
		topicLimit:      49,
		messageBus:      messageBus,
		quitChannel:     quitChannel,
	}
}

func (c *connectionManager) setConnectionLimit(newLimit int) {
	c.connectionLimitMutex.Lock()
	defer c.connectionLimitMutex.Unlock()
	c.connectionLimit = newLimit
}

func (c *connectionManager) setTopicLimit(newLimit int) {
	c.topicLimitMutex.Lock()
	defer c.topicLimitMutex.Unlock()
	c.topicLimit = newLimit
}

func (c *connectionManager) getConnectionLimit() int {
	c.connectionLimitMutex.Lock()
	defer c.connectionLimitMutex.Unlock()
	return c.connectionLimit
}

func (c *connectionManager) getTopicLimit() int {
	c.topicLimitMutex.Lock()
	defer c.topicLimitMutex.Unlock()
	return c.topicLimit
}

func (c *connectionManager) run() {
	for {
		select {
		case <-c.quitChannel:
			for _, conn := range c.connections {
				conn.Disconnect()
			}
			return
			// case <-time.After(1 * time.Second):
			// TODO: Check for orphan topics?
		}
	}
}

func (c *connectionManager) refreshTopic(topic *websocketTopic) {
	topicLimit := c.getTopicLimit()

	for _, conn := range c.connections {
		if conn.numTopics() >= topicLimit {
			continue
		}

		conn.sendListen(topic)
		return
	}

	if len(c.connections) < c.getConnectionLimit() {
		conn := c.addConnection()
		conn.sendListen(topic)
		return
	}

	fmt.Println("[go-twitch-pubsub] connection and topic limit reached")
}

func (c *connectionManager) addConnection() *connection {
	conn := newConnection(c.host, c.messageBus)
	c.connections = append(c.connections, conn)
	go conn.connect()
	return conn
}

func (c *connectionManager) disconnect() {
	for _, conn := range c.connections {
		if !conn.IsConnected() {
			return
		}

		conn.stopReader()

		conn.Disconnect()
	}
}
