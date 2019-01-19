package twitchpubsub

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type websocketTopic struct {
	name      string
	connected bool
	authToken string
}

const (
	reconnectInterval  = 5 * time.Second
	pingInterval       = 4 * time.Minute
	pongDeadlineTime   = 9 * time.Second
	pubsubHost         = "wss://pubsub-edge.twitch.tv"
	writerBufferLength = 100
	readerBufferLength = 100
)

var (
	// ErrNotConnected is returned if an action is attempted to be performed on a Client when it is not connected
	ErrNotConnected = errors.New("go-twitch-pubsub: Not connected")
)

// Client is the client that connects to Twitch's pubsub servers
type Client struct {
	Host string

	// Callbacks
	onModerationAction func(channelID string, data *ModerationAction)
	onBitsEvent        func(channelID string, data *BitsEvent)
	onUnknown          func(bytes []byte)

	connection *websocket.Conn

	connectedMutex sync.Mutex
	connected      bool

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

// NewClient creates a client struct and fills it in with some default values
func NewClient() *Client {
	c := &Client{
		Host: pubsubHost,

		writer:     make(chan []byte, writerBufferLength),
		writerStop: make(chan bool),

		reader:     make(chan []byte, readerBufferLength),
		readerStop: make(chan bool),
	}

	return c
}

// OnModerationAction attaches the given callback to the moderation action event
func (c *Client) OnModerationAction(callback func(channelID string, data *ModerationAction)) {
	c.onModerationAction = callback
}

// OnBitsEvent attaches the given callback to the bits event
func (c *Client) OnBitsEvent(callback func(channelID string, data *BitsEvent)) {
	c.onBitsEvent = callback
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

// Connect starts attempting to connect to the pubsub host
func (c *Client) Connect() error {
	var err error

	c.connection, _, err = websocket.DefaultDialer.Dial(c.Host, nil)
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

func (c *Client) ping() error {
	msg := Base{
		Type: "PING",
	}

	pingTime := time.Now()
	err := c.SendMessage(msg)
	if err != nil {
		return err
	}

	time.AfterFunc(pongDeadlineTime, func() {
		if !c.lastPongWithinLimits(pingTime) {
			fmt.Println("Lost connection, will try to reconnect")
			c.doReconnect = true
			c.onDisconnect()
		}
	})

	return nil
}

func (c *Client) sendMessage(b []byte) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}

	c.writer <- b

	return nil
}

// SendMessage sends a raw message to Twitch's pubsub servers
// Possible errors:
// - If the interface you provide can't be marshalled, a json.Marshal error will be returned
// - If the client is not connected, we can return an `ErrNotConnected` error
func (c *Client) SendMessage(i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}

	return c.sendMessage(b)
}

func (c *Client) startPing() {
	time.AfterFunc(pingInterval, func() {
		_ = c.ping()
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

// Disconnect disconnects from Twitch's pubsub servers and leave the client in an idle state
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

	time.AfterFunc(reconnectInterval, func() {
		c.tryReconnect()
	})
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
			if err := c.parse(payloadBytes); err != nil {
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

func (c *Client) parse(b []byte) (err error) {
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
		fmt.Println("Received unknown message:", string(b))
		return
	}
}

func (c *Client) parseMessage(b []byte) error {
	type message struct {
		Data struct {
			Topic string `json:"topic"`
			// Message is an escaped json string
			Message string `json:"message"`
		} `json:"data"`
	}
	msg := message{}
	if err := json.Unmarshal(b, &msg); err != nil {
		fmt.Println("Error unmarshalling incoming message:", err)
		return nil
	}

	innerMessageBytes := []byte(msg.Data.Message)

	switch getMessageType(msg.Data.Topic) {
	case messageTypeModerationAction:
		d, err := parseModerationAction(innerMessageBytes)
		if err != nil {
			return err
		}
		if c.onModerationAction != nil {
			channelID, err := parseChannelIDFromModerationTopic(msg.Data.Topic)
			if err != nil {
				return err
			}
			c.onModerationAction(channelID, d)
		}
	case messageTypeBitsEvent:
		d, err := parseBitsEvent(innerMessageBytes)
		if err != nil {
			return err
		}
		if c.onBitsEvent != nil {
			channelID, err := parseChannelIDFromBitsTopic(msg.Data.Topic)
			if err != nil {
				return err
			}
			c.onBitsEvent(channelID, d)
		}

	default:
		fallthrough
	case messageTypeUnknown:
		if c.onUnknown != nil {
			c.onUnknown(b)
		}
	}

	return nil
}

func (c *Client) parseResponse(b []byte) error {
	// A "RESPONSE" type message means it's a response to something we sent
	// Most likely, this will be a response to a "LISTEN" message we sent earlier
	// XXX: Right now, we do not attach a nonce to our listens, which means
	// we are unable to identify which message we sent this "RESPONSE" is in response to
	var msg ResponseMessage
	if err := json.Unmarshal(b, &msg); err != nil {
		return err
	}

	if msg.Error != "" {
		fmt.Println("Got an error response to a LISTEN message (don't know which right now):", msg.Error)
	}

	return nil
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

// Listen sends a message to Twitch's pubsub servers telling them we're interested in a specific topic
// Some topics require authentication, and for those you will need to pass a valid authentication token
func (c *Client) Listen(topicName string, authToken string) {
	topic := websocketTopic{topicName, false, authToken}

	c.topicMutex.Lock()
	c.topics = append(c.topics, topic)
	c.topicMutex.Unlock()

	if !c.IsConnected() {
		return
	}

	c.sendListen(topic)
}
