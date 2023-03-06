package twitchpubsub

// Helper functions and structures for twitch AutoMod Queue events

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const autoModQueueEventTopicPrefix = "automod-queue."

// AutoModQueueEvent describes an incoming "AutoMod Queue" action coming from Twitch's PubSub servers
type AutoModQueueEvent struct {
	Message struct {
		ID      string `json:"id"`
		Content struct {
			Text      string `json:"text"`
			Fragments []struct {
				Text    string `json:"text"`
				Automod struct {
					Topics struct {
						Swearing int `json:"swearing"`
					} `json:"topics"`
				} `json:"automod"`
			} `json:"fragments"`
		} `json:"content"`
		Sender struct {
			UserID      string `json:"user_id"`
			Login       string `json:"login"`
			DisplayName string `json:"display_name"`
			ChatColor   string `json:"chat_color"`
		} `json:"sender"`
		SentAt time.Time `json:"sent_at"`
	} `json:"message"`
	ContentClassification struct {
		Category string `json:"category"`
		Level    int    `json:"level"`
	} `json:"content_classification"`
	Status        string `json:"status"`
	ReasonCode    string `json:"reason_code"`
	ResolverID    string `json:"resolver_id"`
	ResolverLogin string `json:"resolver_login"`
}

type outerAutoModQueueEvent struct {
	Data AutoModQueueEvent `json:"data"`
}

func parseAutoModQueueEvent(bytes []byte) (*AutoModQueueEvent, error) {
	data := &outerAutoModQueueEvent{}
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return nil, err
	}

	return &data.Data, nil
}

func parseChannelIDFromAutoModQueueTopic(topic string) (string, error) {
	parts := strings.Split(topic, ".")
	if len(parts) != 3 {
		return "", errors.New("unable to parse channel ID from AutoMod queue topic")
	}

	return parts[2], nil
}

// AutoModQueueEventTopic returns a properly formatted AutoModQueue event topic string with the given channel ID argument
func AutoModQueueEventTopic(modID, channelID string) string {
	const f = `automod-queue.%s.%s`
	return fmt.Sprintf(f, modID, channelID)
}