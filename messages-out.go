package twitchpubsub

const (
	TypeListen = "LISTEN"
)

type ListenData struct {
	Topics    []string `json:"topics"`
	AuthToken string   `json:"auth_token,omitempty"`
}

type Listen struct {
	Type string `json:"type"`

	Nonce string `json:"nonce,omitempty"`

	Data ListenData `json:"data"`
}
