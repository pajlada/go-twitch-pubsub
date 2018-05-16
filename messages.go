package twitch_pubsub

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

type Listen struct {
	Base

	Nonce string `json:"nonce,omitempty"`

	Data ListenData `json:"data"`
}
