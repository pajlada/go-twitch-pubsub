package twitchpubsub

// Helper functions and structures for twitch subscription events
// See examples & structure here: https://dev.twitch.tv/docs/pubsub/#example-channel-subscriptions-event-message

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const subscribeEventTopicPrefix = "channel-subscribe-events-v1."

// SubscribeEvent describes an incoming subscription event on Twitch
type SubscribeEvent struct {
	// ChannelID is the channel that has been subscribed or subgifted to
	ChannelID string `json:"channel_id"`

	// ChannelName is the login name of of the channel that has been subscribed or subgifted to
	ChannelName string `json:"channel_name"`

	// UserID is the ID of the user who subscribed or sent the gift subscription
	// Can be empty if it was an anonymous gift
	UserID string `json:"user_id"`

	// UserName is the login name of the user who subscribed or gifted the subscription
	// Can be empty if it was an anonymous gift
	UserName string `json:"user_name"`

	// DisplayName is the display name of the user who subscribed or gifted the subscription
	// Can be empty if it was an anonymous gift
	DisplayName string `json:"display_name"`

	// Time when the subscription or gift was completed
	Time time.Time `json:"time"`

	// SubPlan is the subscription plan ID (e.g. Prime, 1000, 2000, 3000)
	SubPlan string `json:"sub_plan"`

	// SubPlanName is the subscription plan name
	SubPlanName string `json:"sub_plan_name"`

	// CumulativeMonths is the cumulative number of tenure months of the subscription
	CumulativeMonths int `json:"cumulative_months"`

	// StreakMonths denotes the user's most recent (and contiguous) subscription tenure streak in the channel
	StreakMonths int `json:"streak_months"`

	// Event type associated with the subscription product
	Context string `json:"context"`

	// IsGift denotes whether this subscription was caused by a gift subscription
	IsGift bool `json:"is_gift"`

	SubMessage SubMessage `json:"sub_message"`
}

type Emotes struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	ID    string `json:"id"`
}

type SubMessage struct {
	Message string   `json:"message"`
	Emotes  []Emotes `json:"emotes"`
}

func parseSubscribeEvent(bytes []byte) (*SubscribeEvent, error) {
	data := &SubscribeEvent{}
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func parseChannelIDFromSubscribeTopic(topic string) (string, error) {
	parts := strings.Split(topic, ".")
	if len(parts) != 2 {
		return "", errors.New("unable to parse channel ID from subscribe topic")
	}

	return parts[1], nil
}

func isSubscribeEventTopic(topic string) bool {
	return strings.HasPrefix(topic, subscribeEventTopicPrefix)
}

// SubscribeEventTopic returns a properly formatted subscription event topic string with the given channel ID argument
func SubscribeEventTopic(channelID string) string {
	const f = `channel-subscribe-events-v1.%s`
	return fmt.Sprintf(f, channelID)
}
