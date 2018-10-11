// Helper functions and structures for twitch bits events
package twitchpubsub

import (
	"encoding/json"
	"fmt"
	"time"
)

type BitsEvent struct {
	UserName         string    `json:"user_name"`
	ChannelName      string    `json:"channel_name"`
	UserID           string    `json:"user_id"`
	ChannelID        string    `json:"channel_id"`
	Time             time.Time `json:"time"`
	ChatMessage      string    `json:"chat_message"`
	BitsUsed         int       `json:"bits_used"`
	TotalBitsUsed    int       `json:"total_bits_used"`
	Context          string    `json:"context"`
	BadgeEntitlement struct {
		NewVersion      int `json:"new_version"`
		PreviousVersion int `json:"previous_version"`
	} `json:"badge_entitlement"`
}

func GetBitsEvent(bytes []byte) (*BitsEvent, error) {
	innerData, err := getInnerData(bytes)
	if err != nil {
		return nil, err
	}

	var e BitsEvent
	if err = json.Unmarshal(innerData, &e); err != nil {
		return nil, err
	}

	return &e, nil
}

// Returns a properly formatted bits event topic string with the given channel ID argument
func BitsEventTopic(channelID string) string {
	const f = `channel-bits-events-v1.%s`
	return fmt.Sprintf(f, channelID)
}
