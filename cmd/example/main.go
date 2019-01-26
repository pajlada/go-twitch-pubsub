package main

import (
	"fmt"

	"github.com/pajlada/go-twitch-pubsub"
)

func main() {
	pubsubClient := twitchpubsub.NewClient(twitchpubsub.DefaultHost)

	userID := "82008718"
	channelID := "11148817"

	// OAuth token for userID with chat_login (or chat:read?) scope
	userToken := "dkijfghuidrghuidrgh"

	// Listen to a topic
	pubsubClient.Listen(twitchpubsub.ModerationActionTopic(userID, channelID), userToken)
	pubsubClient.Listen(twitchpubsub.ModerationActionTopic(userID, "93031467"), userToken)
	pubsubClient.Listen(twitchpubsub.ModerationActionTopic(userID, "93031467"), userToken)
	pubsubClient.Listen(twitchpubsub.ModerationActionTopic(userID, "93031467"), userToken)

	// Specify what callback is called when that topic receives a message
	pubsubClient.OnModerationAction(func(channelID string, event *twitchpubsub.ModerationAction) {
		fmt.Println(channelID, event.CreatedBy, event.ModerationAction, "on", event.TargetUserID)
	})

	go pubsubClient.Start()

	c := make(chan bool)
	// select {
	// case <-c:
	// case <-time.After(10 * time.Second):
	// 	pubsubClient.Disconnect()
	// }

	<-c
}
