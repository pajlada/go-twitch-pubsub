// Structures that make up incoming messages from twitch
package twitchpubsub

import (
	"encoding/json"
)

type InnerData struct {
	Data        json.RawMessage `json:"data"`
	Version     string          `json:"version"`
	MessageType string          `json:"message_type"`
	MessageID   string          `json:"message_id"`
}

type BaseData struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

type Message struct {
	Base

	Data BaseData `json:"data"`
}

type ResponseMessage struct {
	Base

	Error string `json:"error"`
	Nonce string `json:"nonce"`
}

func getInnerData(bytes []byte) (json.RawMessage, error) {
	var baseMessage Message
	err := json.Unmarshal(bytes, &baseMessage)
	if err != nil {
		return nil, err
	}

	var innerData InnerData
	if err = json.Unmarshal([]byte(baseMessage.Data.Message), &innerData); err != nil {
		return nil, err
	}

	return innerData.Data, nil
}
