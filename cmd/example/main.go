package main

import (
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
		event, err := twitchpubsub.GetModerationAction(bytes)
		if err != nil {
			return err
		}

		fmt.Println(event.CreatedBy, event.ModerationAction, "on", event.TargetUserID)

		return nil
	})

	err := pubsubClient.Connect()
	if err != nil {
		panic(err)
	}

	c := make(chan bool)
	<-c
}
