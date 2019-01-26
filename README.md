# go-twitch-pubsub
Twitch PubSub library for Go

## Getting Started
```go
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
	userToken := "jdgfkhjkdfhgdfjg"

	// Listen to a topic
	pubsubClient.Listen(twitchpubsub.ModerationActionTopic(userID, channelID), userToken)

	// Specify what callback is called when that topic receives a message
	pubsubClient.OnModerationAction(func(channelID string, event *twitchpubsub.ModerationAction) {
		fmt.Println(event.CreatedBy, event.ModerationAction, "on", event.TargetUserID)
	})

	go pubsubClient.Start()

	c := make(chan bool)
	<-c
}
```
