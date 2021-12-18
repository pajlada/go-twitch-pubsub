package twitchpubsub

import (
	"encoding/json"
	"fmt"
	"strings"
)

// topicMessage defines an incoming pubsub message that has had its type figured out, and parsed according to its type
type topicMessage interface {
	handle(c *Client) error
}

// pubSubMessage defines an incoming pubsub message that hasn't had its type figured out yet
type pubSubMessage struct {
	Type string `json:"type"`
}

type parser interface {
	Parse(topic string, innerMessageBytes []byte) (topicMessage, error)
}

var parsers map[string]parser

func registerParser(topicPrefix string, p parser) {
	if parsers == nil {
		parsers = make(map[string]parser)
	}

	parsers[topicPrefix] = p
}

// readTopicPrefix tries to read a prefix from a topic, given it contains the magical PERIOD character
func readTopicPrefix(topic string) string {
	parts := strings.Split(topic, ".")

	return parts[0]
}

func parseMessage(b []byte) (topicMessage, error) {
	type message struct {
		Data struct {
			Topic string `json:"topic"`
			// Message is an escaped json string
			Message string `json:"message"`
		} `json:"data"`
	}
	msg := message{}
	if err := json.Unmarshal(b, &msg); err != nil {
		return nil, fmt.Errorf("%w: json parsing failed: %s", ErrMalformedMessage, err.Error())
	}

	topic := msg.Data.Topic

	innerMessageBytes := []byte(msg.Data.Message)

	topicPrefix := readTopicPrefix(topic)

	parser, ok := parsers[topicPrefix]
	if !ok {
		return nil, fmt.Errorf("%w: %s/%s", ErrNoParserAvailable, topic, topicPrefix)
	}

	return parser.Parse(topic, innerMessageBytes)
}
