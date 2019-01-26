package twitchpubsub

import (
	"encoding/json"
)

type sharedMessage struct {
	Topic   string
	Message interface{}
}

// InnerData TODO: Refactor
type InnerData struct {
	Data        json.RawMessage `json:"data"`
	Version     string          `json:"version"`
	MessageType string          `json:"message_type"`
	MessageID   string          `json:"message_id"`
}

// BaseData TODO: Refactor
type BaseData struct {
	Topic string `json:"topic"`
	// Message is an escaped json string
	Message string `json:"message"`
}

// Message TODO: Refactor
type Message struct {
	Base

	Data BaseData `json:"data"`
}

// ResponseMessage TODO: Refactor
type ResponseMessage struct {
	Base

	Error string `json:"error"`
	Nonce string `json:"nonce"`
}
