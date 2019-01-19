package twitchpubsub

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const moderationActionTopicPrefix = "chat_moderator_actions."

// ModerationAction describes an incoming "Moderation" action coming from Twitch's PubSub servers
type ModerationAction struct {
	Type             string   `json:"type"`
	ModerationAction string   `json:"moderation_action"`
	Arguments        []string `json:"args"`
	CreatedBy        string   `json:"created_by"`
	CreatedByUserID  string   `json:"created_by_user_id"`
	MsgID            string   `json:"msg_id"`
	TargetUserID     string   `json:"target_user_id"`
}

type outerModerationAction struct {
	Data ModerationAction `json:"data"`
}

func parseModerationAction(bytes []byte) (*ModerationAction, error) {
	data := &outerModerationAction{}
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return nil, err
	}

	return &data.Data, nil
}

func parseChannelIDFromModerationTopic(topic string) (string, error) {
	parts := strings.Split(topic, ".")
	if len(parts) != 3 {
		return "", errors.New("Unable to parse channel ID from moderation topic")
	}

	return parts[2], nil
}

// ModerationActionTopic returns a properly formatted moderation action topic string with the given user and channel ID arguments
func ModerationActionTopic(userID, channelID string) string {
	const f = `chat_moderator_actions.%s.%s`
	return fmt.Sprintf(f, userID, channelID)
}
