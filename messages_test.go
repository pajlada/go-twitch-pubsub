package twitchpubsub

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestReadTopicPrefix(t *testing.T) {
	c := qt.New(t)

	tests := map[string]string{
		"chat_moderator_actions.123.456": "chat_moderator_actions",
		"chat_moderator_actions.456.789": "chat_moderator_actions",
		"topic_without_parameters":       "topic_without_parameters",
		"":                               "",
	}

	for topic, expectedTopicPrefix := range tests {
		topicPrefix := readTopicPrefix(topic)
		c.Assert(topicPrefix, qt.Equals, expectedTopicPrefix)
	}
}

func TestParseMessageGood(t *testing.T) {
	c := qt.New(t)

	tests := []string{
		// ChatModeratorAction User timed out
		`{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.117166826.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"type\":\"chat_login_moderation\",\"moderation_action\":\"timeout\",\"args\":[\"weeb123456\",\"5\",\"reason here\"],\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"created_at\":\"2021-12-18T15:39:38.525054579Z\",\"msg_id\":\"\",\"target_user_id\":\"163915749\",\"target_user_login\":\"\",\"from_automod\":false}}"}}`,

		// ChatModeratorAction  User banned
		`{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.117166826.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"type\":\"chat_login_moderation\",\"moderation_action\":\"ban\",\"args\":[\"weeb123456\",\"reason xd\"],\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"created_at\":\"2021-12-18T15:45:43.448962982Z\",\"msg_id\":\"\",\"target_user_id\":\"163915749\",\"target_user_login\":\"\",\"from_automod\":false}}"}}`,

		// ChatModeratorAction User unbanned
		`{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.117166826.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"type\":\"chat_login_moderation\",\"moderation_action\":\"unban\",\"args\":[\"weeb123456\"],\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"created_at\":\"2021-12-18T15:46:37.786932041Z\",\"msg_id\":\"\",\"target_user_id\":\"163915749\",\"target_user_login\":\"\",\"from_automod\":false}}"}}`,

		// ChatModeratorAction User unmodded
		`{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.117166826.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"type\":\"chat_login_moderation\",\"moderation_action\":\"unmod\",\"args\":[\"ampzyh\"],\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"created_at\":\"2021-12-18T15:47:28.891140838Z\",\"msg_id\":\"\",\"target_user_id\":\"40910607\",\"target_user_login\":\"\",\"from_automod\":false}}"}}`,

		// ChatModeratorAction User modded v1
		`{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.117166826.11148817","message":"{\"type\":\"moderator_added\",\"data\":{\"channel_id\":\"11148817\",\"target_user_id\":\"40910607\",\"moderation_action\":\"mod\",\"target_user_login\":\"ampzyh\",\"created_by_user_id\":\"11148817\",\"created_by\":\"pajlada\"}}"}}`,

		// ChatModeratorAction User modded v2
		`{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.117166826.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"type\":\"chat_login_moderation\",\"moderation_action\":\"mod\",\"args\":[\"ampzyh\"],\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"created_at\":\"2021-12-18T15:47:50.587489942Z\",\"msg_id\":\"\",\"target_user_id\":\"40910607\",\"target_user_login\":\"\",\"from_automod\":false}}"}}`,
	}

	for _, testMessage := range tests {
		tm, err := parseMessage([]byte(testMessage))
		c.Assert(err, qt.IsNil)
		c.Assert(tm, qt.IsNotNil)
	}
}

func TestParseMessageMalformedMessage(t *testing.T) {
	c := qt.New(t)

	tests := []string{
		// ChatModeratorAction: Bad message
		`{"type":MESSAGE","data":{"topic":"chat_moderator_actions.117166826.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"type\":\"chat_login_moderation\",\"moderation_action\":\"mod\",\"args\":[\"ampzyh\"],\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"created_at\":\"2021-12-18T15:47:50.587489942Z\",\"msg_id\":\"\",\"target_user_id\":\"40910607\",\"target_user_login\":\"\",\"from_automod\":false}}"}}`,
	}

	for _, testMessage := range tests {
		tm, err := parseMessage([]byte(testMessage))
		c.Assert(err, qt.ErrorIs, ErrMalformedMessage)
		c.Assert(tm, qt.IsNil)
	}
}

func TestParseMessageMalformedInnerPayload(t *testing.T) {
	c := qt.New(t)

	tests := []string{
		// ChatModeratorAction
		`{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.117166826.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"type\":\"chat_login_moderation\",\"moderation_action\":\"mod\",\"args\":[\"ampzyh\"],\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"created_at\":\"2021-12-18T15:47:50.587489942Z\",\"msg_id\":\"\",\"target_user_id\":\"40910607\",\"target_user_login\":\"\",\"from_automod\":ffalse}}"}}`,
	}

	for _, testMessage := range tests {
		tm, err := parseMessage([]byte(testMessage))
		c.Assert(err, qt.ErrorIs, ErrMalformedInnerPayload)
		c.Assert(tm, qt.IsNil)
	}
}

func TestParseMessageUnhandledTopic(t *testing.T) {
	c := qt.New(t)

	tests := []string{
		`{"type":"MESSAGE","data":{"topic":"no_parser_for_this_topic.117166826.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"type\":\"chat_login_moderation\",\"moderation_action\":\"timeout\",\"args\":[\"weeb123456\",\"5\",\"reason here\"],\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"created_at\":\"2021-12-18T15:39:38.525054579Z\",\"msg_id\":\"\",\"target_user_id\":\"163915749\",\"target_user_login\":\"\",\"from_automod\":false}}"}}`,
	}

	for _, testMessage := range tests {
		tm, err := parseMessage([]byte(testMessage))
		c.Assert(err, qt.ErrorIs, ErrNoParserAvailable)
		c.Assert(tm, qt.IsNil)
	}
}
