package twitchpubsub

import (
	"errors"
	"sync"
	"time"
)

const (
	reconnectInterval      = 5 * time.Second
	pingInterval           = 4 * time.Minute
	pongDeadlineTime       = 9 * time.Second
	writerBufferLength     = 100
	readerBufferLength     = 100
	messageBusBufferLength = 50

	// maximum number of connections to open
	defaultConnectionLimit = 10

	// maximum number of topics one connection can listen to
	defaultTopicLimit = 50
)

var (
	// ErrNotConnected is returned if an action is attempted to be performed on a Client when it is not connected
	ErrNotConnected = errors.New("go-twitch-pubsub: Not connected")

	// ErrDisconnectedByUser is returned from Connect after the user calls Disconnect()
	ErrDisconnectedByUser = errors.New("go-twitch-pubsub: Disconnected by user")

	ErrMissingCallback = errors.New("missing callback for topic")

	ErrNoParserAvailable = errors.New("no parser available for topic")

	ErrMalformedInnerPayload = errors.New("malformed inner payload")

	ErrMalformedMessage = errors.New("malformed message")

	// DefaultHost is the default host to connect to Twitch's pubsub servers
	DefaultHost = "wss://pubsub-edge.twitch.tv"
)

type messageBusType chan topicMessage

// Client is the client that connects to Twitch's pubsub servers
type Client struct {
	// onChatModeratorAction is the callback for the chat_moderator_actions topic
	// The implementation for this can be found in the chat_moderator_actions.go file
	onChatModerationAction func(channelID string, data *ChatModeratorAction)

	// onBitsEvent is the ca
	onBitsEvent func(channelID string, data *BitsEvent)

	connectionManager *connectionManager

	topics *topicManager

	messageBus messageBusType

	quitChannel chan struct{}
}

// NewClient creates a client struct and fills it in with some default values
func NewClient(host string) *Client {
	c := &Client{
		messageBus:  make(messageBusType, messageBusBufferLength),
		quitChannel: make(chan struct{}),

		topics: newTopicManager(),
	}

	c.connectionManager = &connectionManager{
		host: host,

		connectionLimit:      defaultConnectionLimit,
		connectionLimitMutex: &sync.RWMutex{},

		topicLimit:      defaultTopicLimit,
		topicLimitMutex: &sync.RWMutex{},

		messageBus:  c.messageBus,
		quitChannel: c.quitChannel,
	}

	return c
}

func (c *Client) SetConnectionLimit(connectionLimit int) {
	c.connectionManager.setConnectionLimit(connectionLimit)
}

func (c *Client) SetTopicLimit(topicLimit int) {
	c.connectionManager.setTopicLimit(topicLimit)
}

// OnBitsEvent attaches the given callback to the bits event
func (c *Client) OnBitsEvent(callback func(channelID string, data *BitsEvent)) {
	c.onBitsEvent = callback
}

// Connect starts attempting to connect to the pubsub host
func (c *Client) Start() error {
	go c.connectionManager.run()

	for {
		select {
		case iMsg := <-c.messageBus:
			if err := iMsg.handle(c); err != nil {
				return err
			}

			// switch msg := iMsg.Message.(type) {
			// case *ChatModeratorAction:
			// 	if err := c.handleChatModeratorActionsMessage(iMsg.Topic, msg); err != nil {
			// 		return err
			// 	}

			// case *BitsEvent:
			// 	channelID, err := parseChannelIDFromBitsTopic(iMsg.Topic)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	c.onBitsEvent(channelID, msg)

			// default:
			// 	log.Println("unknown message in message bus")
			// }
		case <-c.quitChannel:
			return ErrDisconnectedByUser
		}
	}
}

// Disconnect disconnects from Twitch's pubsub servers and leave the client in an idle state
func (c *Client) Disconnect() {
	c.connectionManager.disconnect()

	close(c.quitChannel)
}

// Listen sends a message to Twitch's pubsub servers telling them we're interested in a specific topic
// Some topics require authentication, and for those you will need to pass a valid authentication token
func (c *Client) listen(topicName string, authToken string) {
	topic := newTopic(topicName, authToken)

	if !c.topics.Add(topic) {
		// We were already subscribed to this topic
		return
	}

	c.connectionManager.refreshTopic(topic)
}
