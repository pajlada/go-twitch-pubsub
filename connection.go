package twitchpubsub

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var connectionDialer *websocket.Dialer

func setConnectionDialer(d *websocket.Dialer) {
	connectionDialer = d
}

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

	nonceCounter uint64

	// numConnects gets incremented at the start of each connect call
	// this is only used for tests
	numConnects atomic.Uint64

	reconnectInterval time.Duration
	pingInterval      time.Duration
	pongDeadlineTime  time.Duration
}

func newConnection(host string, messageBus messageBusType) *connection {
	return &connection{
		host: host,

		writer:     make(chan []byte, writerBufferLength),
		writerStop: make(chan bool),

		reader:     make(chan []byte, readerBufferLength),
		readerStop: make(chan bool),

		messageBus: messageBus,

		reconnectInterval: defaultReconnectInterval,
		pingInterval:      defaultPingInterval,
		pongDeadlineTime:  defaultPongDeadlineTime,
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
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Println("[go-twitch-pubsub]: Unexpected close error:", err)
				}
				c.readerStop <- true
				return
			}

			if messageType == websocket.TextMessage {
				c.reader <- payloadBytes
			}
		}
	}()

	pingTime := time.Now()
	pingTicker := time.NewTicker(c.pingInterval)
	pongCheckTimer := time.NewTimer(c.pongDeadlineTime)
	pongCheckTimer.Stop()

	for {
		select {
		case payloadBytes := <-c.reader:
			if err := c.parse(payloadBytes); err != nil {
				fmt.Println("Error parsing received websocket message:", err)
			}

		case <-pingTicker.C:
			if err := c.ping(); err != nil {
				fmt.Println("[go-twitch-pubsub] Error sending ping:", err)
				return
			}
			pingTime = time.Now()
			pongCheckTimer.Reset(c.pongDeadlineTime)

		case <-pongCheckTimer.C:
			if !c.lastPongWithinLimits(pingTime) {
				fmt.Println("[go-twitch-pubsub] Lost connection, will try to reconnect")
				return
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

	lastPongDiff := c.lastPong.Sub(pingTime)
	return lastPongDiff >= 0*time.Second && lastPongDiff < c.pongDeadlineTime
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

	err := c.writeMessage(msg)
	if err != nil {
		return err
	}

	return nil
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
	c.numConnects.Add(1)
	var err error
	if connectionDialer == nil {
		c.wsConn, _, err = websocket.DefaultDialer.Dial(c.host, nil)
	} else {
		c.wsConn, _, err = connectionDialer.Dial(c.host, nil)
	}
	if err != nil {
		c.doReconnect = true
		c.onDisconnect()
		return err
	}

	c.setConnected(true)

	go c.startReader()
	go c.startWriter()

	return nil
}

func (c *connection) onDisconnect() {
	if !c.doReconnect {
		// Do full closes
		return
	}

	time.AfterFunc(c.reconnectInterval, func() {
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
	case messageTypePointsEvent:
		d, err := parsePointsEvent(innerMessageBytes)
		if err != nil {
			return err
		}
		c.messageBus <- sharedMessage{
			Topic:   msg.Data.Topic,
			Message: d,
		}
	case messageTypeAutoModQueueEvent:
		d, err := parseAutoModQueueEvent(innerMessageBytes)
		if err != nil {
			return err
		}
		c.messageBus <- sharedMessage{
			Topic:   msg.Data.Topic,
			Message: d,
		}
	case messageTypeWhisperEvent:
		d, err := parseWhisperEvent(innerMessageBytes)
		if err != nil {
			return err
		}
		c.messageBus <- sharedMessage{
			Topic:   msg.Data.Topic,
			Message: d,
		}
	case messageTypeSubscribeEvent:
		d, err := parseSubscribeEvent(innerMessageBytes)
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

func (c *connection) getNonce() string {
	v := atomic.AddUint64(&c.nonceCounter, 1)
	return strconv.FormatUint(v, 10)
}

func (c *connection) sendListen(topic *websocketTopic) {
	nonce := c.getNonce()
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
