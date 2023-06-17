package twitchpubsub

// Helper functions and structures for twitch bits events

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const bitsEventTopicPrefix = "channel-bits-events-v1."

// BitsEvent describes an incoming "Bit" action coming from Twitch's PubSub servers
type BitsEvent struct {
	// UserName is the bit sender's login name
	UserName string `json:"user_name"`
	// UserID is the bit sender's ID
	UserID string `json:"user_id"`

	// ChannelName is the bit-recipient user's login name
	ChannelName string `json:"channel_name"`
	// ChannelID is the bit-recipient user's ID
	ChannelID string `json:"channel_id"`

	// Time the bits were sent
	Time time.Time `json:"time"`

	// ChatMessage sent along with the bits
	ChatMessage string `json:"chat_message"`

	// BitsUsed is the number of bits sent in this message
	BitsUsed int `json:"bits_used"`

	// TotalBitsUsed is the total number of bits the user has sent
	TotalBitsUsed int `json:"total_bits_used"`

	Context string `json:"context"`

	BadgeEntitlement struct {
		NewVersion      int `json:"new_version"`
		PreviousVersion int `json:"previous_version"`
	} `json:"badge_entitlement"`
}

type outerBitsEvent struct {
	Data BitsEvent `json:"data"`
}

func parseBitsEvent(bytes []byte) (*BitsEvent, error) {
	data := &outerBitsEvent{}
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return nil, err
	}

	return &data.Data, nil
}

func parseChannelIDFromBitsTopic(topic string) (string, error) {
	parts := strings.Split(topic, ".")
	if len(parts) != 2 {
		return "", errors.New("unable to parse channel ID from bits topic")
	}

	return parts[1], nil
}

func isBitsEventTopic(topic string) bool {
	return strings.HasPrefix(topic, bitsEventTopicPrefix)
}

// BitsEventTopic returns a properly formatted bits event topic string with the given channel ID argument
func BitsEventTopic(channelID string) string {
	const f = `channel-bits-events-v1.%s`
	return fmt.Sprintf(f, channelID)
}
