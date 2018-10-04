package twitchpubsub

import "fmt"

const (
	TypeListen = "LISTEN"
)

type Base struct {
	Type string `json:"type"`
}

type BaseData struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

type Message struct {
	Base

	Data BaseData `json:"data"`
}

type ListenData struct {
	Topics    []string `json:"topics"`
	AuthToken string   `json:"auth_token"`
}

type TimeoutData struct {
	Data struct {
		Type             string   `json:"type"`
		ModerationAction string   `json:"moderation_action"`
		Arguments        []string `json:"args"`
		CreatedBy        string   `json:"created_by"`
		CreatedByUserID  string   `json:"created_by_user_id"`
		MsgID            string   `json:"msg_id"`
		TargetUserID     string   `json:"target_user_id"`
	} `json:"data"`
}

// Returns a properly formatted moderation action topic string with the given user and channel ID arguments
func ModerationActionTopic(userID, channelID string) string {
	const f = `chat_moderator_actions.%s.%s`
	return fmt.Sprintf(f, userID, channelID)
}

type Listen struct {
	Base

	Nonce string `json:"nonce,omitempty"`

	Data ListenData `json:"data"`
}
