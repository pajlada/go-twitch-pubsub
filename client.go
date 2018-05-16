package twitch_pubsub

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// SubscribeCallback defines a callback that is called when a Publish method matches the subscription topic
type SubscribeCallback func(rawMessage []byte) error
type subscribeCallbackMap map[string]SubscribeCallback

type websocketTopic struct {
	name      string
	connected bool
	authToken string
}

const reconnectInterval = 5 * time.Second
const pingInterval = 4 * time.Minute
const pongDeadlineTime = 9 * time.Second
const pubsubHost = "wss://pubsub-edge.twitch.tv"
const writerBufferLength = 100
const readerBufferLength = 100

var ErrNotConnected = errors.New("go-twitch-pubsub: Not connected")

type Client struct {
	Host string

	connection *websocket.Conn

	connectedMutex sync.Mutex
	connected      bool

	callbacks subscribeCallbackMap

	topicMutex sync.Mutex
	topics     []websocketTopic

	writer     chan []byte
	writerStop chan bool

	reader     chan []byte
	readerStop chan bool

	doReconnect bool

	pongMutex sync.Mutex
	lastPong  time.Time
}

func NewClient() *Client {
	c := &Client{
		Host: pubsubHost,
	}

	c.callbacks = make(subscribeCallbackMap)

	c.writer = make(chan []byte, writerBufferLength)
	c.writerStop = make(chan bool)

	c.reader = make(chan []byte, readerBufferLength)
	c.readerStop = make(chan bool)

	return c
}

func (c *Client) onPong() {
	c.pongMutex.Lock()
	c.lastPong = time.Now()
	c.pongMutex.Unlock()
}

func (c *Client) lastPongWithinLimits(pingTime time.Time) bool {
	c.pongMutex.Lock()
	defer c.pongMutex.Unlock()

	return c.lastPong.Sub(pingTime) < pongDeadlineTime
}

func (c *Client) Connect() error {
	var err error

	c.connection, _, err = websocket.DefaultDialer.Dial(pubsubHost, nil)
	if err != nil {
		c.doReconnect = true
		c.onDisconnect()
		return err
	}

	c.setConnected(true)

	go c.startReader()
	go c.startWriter()

	c.subscribeToTopics()

	c.onConnected()

	return nil
}

func (c *Client) onConnected() {
	go c.startPing()
}

func (c *Client) Ping() {
	msg := Base{
		Type: "PING",
	}

	pingTime := time.Now()
	c.SendMessage(msg)

	time.AfterFunc(pongDeadlineTime, func() {
		if !c.lastPongWithinLimits(pingTime) {
			fmt.Println("Lost connection, will try to reconnect")
			c.doReconnect = true
			c.onDisconnect()
		}
	})
}

func (c *Client) sendMessage(b []byte) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}

	c.writer <- b

	return nil
}

func (c *Client) SendMessage(i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return errors.New("SendInterface JSON Error")
	}

	if err = c.sendMessage(b); err != nil {
		return errors.New("SendMessage sendMessage error")
	}

	return nil
}

func (c *Client) startPing() {
	time.AfterFunc(pingInterval, func() {
		c.Ping()
		c.startPing()
	})
}

// IsConnected returns the current connection state
func (c *Client) IsConnected() bool {
	c.connectedMutex.Lock()
	defer c.connectedMutex.Unlock()

	return c.connected
}

func (c *Client) setConnected(newConnectedState bool) {
	c.connectedMutex.Lock()
	c.connected = newConnectedState
	c.connectedMutex.Unlock()
}

func (c *Client) stopWriter() {
	c.writerStop <- true

	for len(c.writer) > 0 {
		<-c.writer
	}
}

func (c *Client) stopReader() {
	c.readerStop <- true

	for len(c.reader) > 0 {
		<-c.reader
	}
}

func (c *Client) Disconnect() {
	if !c.IsConnected() {
		return
	}

	c.stopReader()
	// when stopReader has finished, it will stop the writer

	c.topicMutex.Lock()
	for _, topic := range c.topics {
		topic.connected = false
	}
	c.topicMutex.Unlock()

	c.connection.Close()
}

func (c *Client) onDisconnect() {
	if !c.doReconnect {
		return
	}

	c.tryReconnect()
}

func (c *Client) tryReconnect() error {
	return c.Connect()
}

func (c *Client) startReader() {
	defer func() {
		c.doReconnect = true
		c.setConnected(false)
		c.stopWriter()
		c.connection.Close()
		c.onDisconnect()
	}()

	// Read
	go func() {
		for {
			messageType, payloadBytes, err := c.connection.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					fmt.Println("Gracefully disconnected")
				} else {
					fmt.Println("Unknown error in read channel:", err)
				}

				c.readerStop <- true
				return
			}

			if messageType == websocket.TextMessage {
				c.reader <- payloadBytes
			} else {
				fmt.Println("Unhandled Message type:", messageType, "with payload", string(payloadBytes))
			}
		}
	}()

	for {
		select {
		case payloadBytes := <-c.reader:
			if err := c.parseMessage(payloadBytes); err != nil {
				fmt.Println("Error parsing received websocket message:", err)
			}

		case <-c.readerStop:
			return
		}
	}
}

func (c *Client) startWriter() {
	for {
		select {
		case payload := <-c.writer:
			c.connection.WriteMessage(websocket.TextMessage, payload)

		case <-c.writerStop:
			return
		}
	}
}

func (c *Client) parseMessage(b []byte) (err error) {
	baseMsg := Base{}
	err = json.Unmarshal(b, &baseMsg)
	if err != nil {
		return
	}

	switch baseMsg.Type {
	case "PONG":
		c.onPong()

	case "MESSAGE":
		msg := Message{}
		err = json.Unmarshal(b, &msg)
		if err != nil {
			fmt.Println("Error unmarshalling message:", err)
			return
		}
		if cb, ok := c.callbacks[msg.Data.Topic]; ok {
			err = cb(b)
			if err != nil {
				fmt.Println("Error calling message callback:", err)
				return
			}

			return nil
		}

	default:
		fmt.Println("Received unknown message:", string(b))
	}

	return
}

func (c *Client) subscribeToTopics() {
	if !c.IsConnected() {
		return
	}

	c.topicMutex.Lock()
	for _, topic := range c.topics {
		if topic.connected {
			continue
		}

		c.sendListen(topic)
	}
	c.topicMutex.Unlock()
}

func (c *Client) sendListen(topic websocketTopic) error {
	msg := Listen{
		Base: Base{
			Type: TypeListen,
		},
		Data: ListenData{
			Topics:    []string{topic.name},
			AuthToken: topic.authToken,
		},
	}

	return c.SendMessage(msg)
}

func (c *Client) Listen(topicName string, authToken string, cb SubscribeCallback) {
	c.callbacks[topicName] = cb

	topic := websocketTopic{topicName, false, authToken}

	c.topicMutex.Lock()
	c.topics = append(c.topics, topic)
	c.topicMutex.Unlock()

	if !c.IsConnected() {
		return
	}

	c.sendListen(topic)
}
