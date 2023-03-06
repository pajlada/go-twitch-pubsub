package twitchpubsub

// Helper functions and structures for twitch whisper events

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const whisperEventTopicPrefix = "whispers."

// WhisperEvent describes an incoming whisper coming from Twitch's PubSub servers
type WhisperEvent struct {
	Type string `json:"type"`
	Data struct {
		ID int `json:"id"`
	} `json:"data"`
	ThreadID string `json:"thread_id"`
	Body     string `json:"body"`
	SentTs   int    `json:"sent_ts"`
	FromID   int    `json:"from_id"`
	Tags     struct {
		Login       string        `json:"login"`
		DisplayName string        `json:"display_name"`
		Color       string        `json:"color"`
		Emotes      []interface{} `json:"emotes"`
		Badges      []struct {
			ID      string `json:"id"`
			Version string `json:"version"`
		} `json:"badges"`
	} `json:"tags"`
	Recipient struct {
		ID          int           `json:"id"`
		Username    string        `json:"username"`
		DisplayName string        `json:"display_name"`
		Color       string        `json:"color"`
		Badges      []interface{} `json:"badges"`
	} `json:"recipient"`
	Nonce string `json:"nonce"`
}

func parseWhisperEvent(bytes []byte) (*WhisperEvent, error) {
	data := &WhisperEvent{}
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func parseUserIDFromWhisperTopic(topic string) (string, error) {
	parts := strings.Split(topic, ".")
	if len(parts) != 3 {
		return "", errors.New("unable to parse channel ID from whisper topic")
	}

	return parts[2], nil
}

// WhisperEventTopic returns a properly formatted whisper event topic string with the given userID ID argument
func WhisperEventTopic(userID string) string {
	const f = `whispers.%s`
	return fmt.Sprintf(f, userID)
}