package twitchpubsub

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestParseChannelIDFromModerationTopic(t *testing.T) {
	goodTopics := map[string]string{
		"chat_moderator_actions.123.456": "456",
	}

	badTopics := map[string]string{
		"chat_moderator_actions.123456": "",
	}

	c := qt.New(t)

	for topic, expectedChannelID := range goodTopics {
		channelID, err := parseChannelIDFromChatModeratorActionsTopic(topic)
		c.Assert(err, qt.IsNil)
		c.Assert(channelID, qt.Equals, expectedChannelID)
	}

	for topic := range badTopics {
		channelID, err := parseChannelIDFromChatModeratorActionsTopic(topic)
		c.Assert(err, qt.ErrorIs, ErrMalformedChatModeratorActionsTopic)
		c.Assert(channelID, qt.Equals, "")
	}
}

func TestCreateChatModeratorActionsTopic(t *testing.T) {
	type tt struct {
		userID        string
		channelID     string
		expectedTopic string
	}

	tests := []tt{
		{
			userID:        "123",
			channelID:     "456",
			expectedTopic: "chat_moderator_actions.123.456",
		},
		{
			userID:        "456",
			channelID:     "789",
			expectedTopic: "chat_moderator_actions.456.789",
		},
	}

	c := qt.New(t)

	for _, test := range tests {
		topic := createTopicChatModeratorActions(test.userID, test.channelID)

		c.Assert(topic, qt.Equals, test.expectedTopic)
	}
}
