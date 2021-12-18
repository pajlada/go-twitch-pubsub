package main

import (
	"fmt"
	"os"

	twitchpubsub "github.com/pajlada/go-twitch-pubsub"
)

func main() {
	pubsubClient := twitchpubsub.NewClient(twitchpubsub.DefaultHost)

	userID := "117166826"
	channelID := "11148817"

	// OAuth token for userID with chat_login (or chat:read?) scope
	userToken := os.Getenv("GO_TWITCH_PUBSUB_USER_TOKEN")

	// Listen to a topic
	pubsubClient.ListenChatModeratorActions(userID, channelID, userToken)
	// pubsubClient.ListenChatModeratorActions(userID, "93031467", userToken)
	// pubsubClient.ListenChatModeratorActions(userID, "93031467", userToken)
	// pubsubClient.ListenChatModeratorActions(userID, "93031467", userToken)

	// Specify what callback is called when that topic receives a message
	pubsubClient.OnChatModeratorAction(func(channelID string, event *twitchpubsub.ChatModeratorAction) {
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
