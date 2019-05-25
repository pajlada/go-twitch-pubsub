package twitchpubsub

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pajbot/utils"
)

type connection struct {
	host string

	wsConn *websocket.Conn

	connectedMutex sync.Mutex
	connected      bool

	writer     chan []byte
	writerStop chan bool

	reader     chan []byte
	readerStop chan bool

	pongMutex sync.Mutex
	lastPong  time.Time

	messageBus chan sharedMessage

	doReconnect bool

	topics []*websocketTopic
}

func newConnection(host string, messageBus messageBusType) *connection {
	return &connection{
		host: host,

		writer:     make(chan []byte, writerBufferLength),
		writerStop: make(chan bool),

		reader:     make(chan []byte, readerBufferLength),
		readerStop: make(chan bool),

		messageBus: messageBus,
	}
}

func (c *connection) startReader() {
	defer func() {
		c.doReconnect = true
		c.setConnected(false)
		c.stopWriter()
		c.wsConn.Close()
		c.onDisconnect()
	}()

	// Read
	go func() {
		for {
			messageType, payloadBytes, err := c.wsConn.ReadMessage()
			if err != nil {
				c.readerStop <- true
				return
			}

			if messageType == websocket.TextMessage {
				c.reader <- payloadBytes
			}
		}
	}()

	for {
		select {
		case payloadBytes := <-c.reader:
			if err := c.parse(payloadBytes); err != nil {
				fmt.Println("Error parsing received websocket message:", err)
			}

		case <-c.readerStop:
			return
		}
	}
}

func (c *connection) onPong() {
	c.pongMutex.Lock()
	c.lastPong = time.Now()
	c.pongMutex.Unlock()
}

func (c *connection) lastPongWithinLimits(pingTime time.Time) bool {
	c.pongMutex.Lock()
	defer c.pongMutex.Unlock()

	return c.lastPong.Sub(pingTime) < pongDeadlineTime
}

func (c *connection) writeMessage(msg interface{}) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.writer <- b

	return nil
}

func (c *connection) ping() error {
	msg := &Base{
		Type: "PING",
	}

	pingTime := time.Now()
	err := c.writeMessage(msg)
	if err != nil {
		return err
	}

	time.AfterFunc(pongDeadlineTime, func() {
		if !c.lastPongWithinLimits(pingTime) {
			fmt.Println("[go-twitch-pubsub] Lost connection, will try to reconnect")
			c.doReconnect = true
			c.onDisconnect()
		}
	})

	return nil
}

func (c *connection) startPing() {
	time.AfterFunc(pingInterval, func() {
		_ = c.ping()
		c.startPing()
	})
}

func (c *connection) setConnected(newConnectedState bool) {
	c.connectedMutex.Lock()
	c.connected = newConnectedState
	c.connectedMutex.Unlock()
}

func (c *connection) stopWriter() {
	c.writerStop <- true

	for len(c.writer) > 0 {
		<-c.writer
	}
}

// IsConnected returns the current connection state
func (c *connection) IsConnected() bool {
	c.connectedMutex.Lock()
	defer c.connectedMutex.Unlock()

	return c.connected
}

func (c *connection) Disconnect() {
	c.stopReader()
	// when stopReader has finished, it will stop the writer

	c.wsConn.Close()
}

func (c *connection) stopReader() {
	c.readerStop <- true

	for len(c.reader) > 0 {
		<-c.reader
	}
}

func (c *connection) connect() error {
	var err error
	c.wsConn, _, err = websocket.DefaultDialer.Dial(c.host, nil)
	if err != nil {
		c.doReconnect = true
		c.onDisconnect()
		return err
	}

	c.setConnected(true)

	go c.startReader()
	go c.startWriter()

	c.onConnected()

	return nil
}

func (c *connection) onConnected() {
	go c.startPing()
}

func (c *connection) onDisconnect() {
	if !c.doReconnect {
		// Do full closes
		return
	}

	time.AfterFunc(reconnectInterval, func() {
		c.tryReconnect()
	})
}

func (c *connection) tryReconnect() error {
	return c.connect()
}

func (c *connection) startWriter() {
	for {
		select {
		case payload := <-c.writer:
			c.wsConn.WriteMessage(websocket.TextMessage, payload)

		case <-c.writerStop:
			return
		}
	}
}

func (c *connection) parse(b []byte) (err error) {
	baseMsg := Base{}
	err = json.Unmarshal(b, &baseMsg)
	if err != nil {
		return
	}

	switch baseMsg.Type {
	case "PONG":
		c.onPong()
		return

	case "MESSAGE":
		return c.parseMessage(b)

	case "RESPONSE":
		return c.parseResponse(b)

	default:
		// fmt.Println("Received unknown message:", string(b))
		return
	}
}

func (c *connection) parseMessage(b []byte) error {
	type message struct {
		Data struct {
			Topic string `json:"topic"`
			// Message is an escaped json string
			Message string `json:"message"`
		} `json:"data"`
	}
	msg := message{}
	if err := json.Unmarshal(b, &msg); err != nil {
		fmt.Println("[go-twitch-pubsub] Error unmarshalling incoming message:", err)
		return nil
	}

	innerMessageBytes := []byte(msg.Data.Message)

	switch getMessageType(msg.Data.Topic) {
	case messageTypeModerationAction:
		d, err := parseModerationAction(innerMessageBytes)
		if err != nil {
			return err
		}
		c.messageBus <- sharedMessage{
			Topic:   msg.Data.Topic,
			Message: d,
		}
	case messageTypeBitsEvent:
		d, err := parseBitsEvent(innerMessageBytes)
		if err != nil {
			return err
		}
		c.messageBus <- sharedMessage{
			Topic:   msg.Data.Topic,
			Message: d,
		}

	default:
		fallthrough
	case messageTypeUnknown:
		// This can be used while implementing new message types
	}

	return nil
}

func (c *connection) sendListen(topic *websocketTopic) {
	nonce, _ := utils.GenerateRandomString(32)
	msg := Listen{
		Base: Base{
			Type: TypeListen,
		},
		Nonce: nonce,
		Data: ListenData{
			Topics:    []string{topic.name},
			AuthToken: topic.authToken,
		},
	}

	topic.nonce = nonce

	c.topics = append(c.topics, topic)

	c.writeMessage(msg)
}

func (c *connection) parseResponse(b []byte) error {
	// A "RESPONSE" type message means it's a response to something we sent
	// Most likely, this will be a response to a "LISTEN" message we sent earlier
	var msg ResponseMessage
	if err := json.Unmarshal(b, &msg); err != nil {
		return err
	}

	if msg.Error == "" {
		for _, topic := range c.topics {
			if topic.nonce == msg.Nonce {
				topic.connected = true
				return nil
			}
		}
	} else {
		for _, topic := range c.topics {
			if topic.nonce == msg.Nonce {
				fmt.Println("[go-twitch-pubsub] Error connecting to", topic.hash)
				topic.connected = true
				return nil
			}
		}
	}

	return nil
}

func (c *connection) numTopics() int {
	return len(c.topics)
}
