package twitchpubsub

import "encoding/json"

type outerMessage struct {
	Data struct {
		Topic string `json:"topic"`
		// Message is an escaped json string
		Message string `json:"message"`
	} `json:"data"`
}

func parseOuterMessage(b []byte) (*outerMessage, error) {
	msg := outerMessage{}
	if err := json.Unmarshal(b, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
