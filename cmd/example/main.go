package main

import (
	"encoding/json"
	"fmt"

	"github.com/pajlada/go-twitch-pubsub"
)

func main() {
	pubsubClient := twitchpubsub.NewClient()

	userID := "117166826"
	channelID := "11148817"

	// OAuth token for userID with chat_login (or chat:read?) scope
	userToken := "abcdef123456"

	pubsubClient.Listen(twitchpubsub.ModerationActionTopic(userID, channelID), userToken, func(bytes []byte) error {
		var baseMsg twitchpubsub.Message
		err := json.Unmarshal(bytes, &baseMsg)
		if err != nil {
			return err
		}

		var msg twitchpubsub.TimeoutData
		err = json.Unmarshal([]byte(baseMsg.Data.Message), &msg)
		if err != nil {
			return err
		}

		fmt.Println(msg.Data.CreatedBy, msg.Data.ModerationAction, "on", msg.Data.TargetUserID)

		return nil
	})

	err := pubsubClient.Connect()
	if err != nil {
		panic(err)
	}

	c := make(chan bool)
	<-c
}
