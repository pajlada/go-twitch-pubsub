package twitchpubsub

import (
	"errors"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestParseModerationEvent(t *testing.T) {
	c := qt.New(t)

	type testCase struct {
		label            string
		input            string
		isSubscribeMsg   bool
		expected         *ModerationAction
		expectedErr      error
		expectedOuterErr error
	}

	testCases := []testCase{
		{
			label:          "Timeout without reason",
			input:          `{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.11148817.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"args\":[\"69420\",\"1\",\"\"],\"created_at\":\"2023-06-11T09:58:52.152684347Z\",\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"moderation_action\":\"timeout\",\"target_user_id\":\"25497681\",\"target_user_login\":\"69420\",\"type\":\"chat_login_moderation\"}}"}}`,
			isSubscribeMsg: true,
			expected: &ModerationAction{
				Type:             "chat_login_moderation",
				ModerationAction: "timeout",
				Arguments:        []string{"69420", "1", ""},
				CreatedBy:        "pajlada",
				CreatedByUserID:  "11148817",
				MsgID:            "",
				TargetUserID:     "25497681",
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},
		{
			label:          "Timeout with reason",
			input:          `{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.11148817.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"args\":[\"doge41732\",\"5\",\"This is the reason for the timeout\"],\"created_at\":\"2023-06-17T15:04:31.20928599Z\",\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"moderation_action\":\"timeout\",\"target_user_id\":\"115117172\",\"target_user_login\":\"doge41732\",\"type\":\"chat_login_moderation\"}}"}}`,
			isSubscribeMsg: true,
			expected: &ModerationAction{
				Type:             "chat_login_moderation",
				ModerationAction: "timeout",
				Arguments:        []string{"doge41732", "5", "This is the reason for the timeout"},
				CreatedBy:        "pajlada",
				CreatedByUserID:  "11148817",
				MsgID:            "",
				TargetUserID:     "115117172",
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},
		{
			label:          "Deleted message",
			input:          `{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.11148817.11148817","message":"{\"type\":\"moderation_action\",\"data\":{\"type\":\"chat_login_moderation\",\"moderation_action\":\"delete\",\"args\":[\"slurps\",\"4HEad 👍\",\"31197cd8-d5da-4deb-a146-3d8b5115518a\"],\"created_by\":\"pajlada\",\"created_by_user_id\":\"11148817\",\"created_at\":\"2023-06-17T15:07:17.679767013Z\",\"msg_id\":\"\",\"target_user_id\":\"133077169\",\"target_user_login\":\"\",\"from_automod\":false}}"}}`,
			isSubscribeMsg: true,
			expected: &ModerationAction{
				Type:             "chat_login_moderation",
				ModerationAction: "delete",
				Arguments: []string{
					"slurps",
					"4HEad 👍",
					"31197cd8-d5da-4deb-a146-3d8b5115518a",
				},
				CreatedBy:       "pajlada",
				CreatedByUserID: "11148817",
				MsgID:           "",
				TargetUserID:    "133077169",
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},

		{
			label:            "Invalid message JSON",
			input:            `{"type":"MESSAGE","data":{"topic":"chat_moderator_actions.11148817.11148817","message":"{forsen}"}}`,
			isSubscribeMsg:   true,
			expected:         nil,
			expectedErr:      errors.New("invalid character 'f' looking for beginning of object key string"),
			expectedOuterErr: nil,
		},
	}

	for _, testCase := range testCases {
		c.Run(testCase.label, func(c *qt.C) {
			outerMessage, err := parseOuterMessage([]byte(testCase.input))
			c.Assert(err, qt.Equals, testCase.expectedOuterErr)
			c.Assert(isModerationActionTopic(outerMessage.Data.Topic), qt.Equals, testCase.isSubscribeMsg)

			if testCase.isSubscribeMsg {
				// Only test parsing if we expect the input message to be an actual subscribe message
				innerMessageBytes := []byte(outerMessage.Data.Message)
				actual, err := parseModerationAction([]byte(innerMessageBytes))

				if testCase.expectedErr == nil {
					c.Assert(err, qt.IsNil)
				} else {
					c.Assert(err, qt.ErrorMatches, testCase.expectedErr.Error())
				}
				c.Assert(actual, qt.DeepEquals, testCase.expected)
			}
		})
	}
}

func TestCreateModerationTopic(t *testing.T) {
	c := qt.New(t)

	type testCase struct {
		label          string
		inputUserID    string
		inputChannelID string
		expected       string
	}

	testCases := []testCase{
		{
			label:          "Standard",
			inputUserID:    "123",
			inputChannelID: "456",
			expected:       "chat_moderator_actions.123.456",
		},
		{
			label:          "Bad",
			inputUserID:    "forsen",
			inputChannelID: "456",
			expected:       "chat_moderator_actions.forsen.456",
		},
		{
			label:          "Bad 2",
			inputUserID:    "",
			inputChannelID: "456",
			expected:       "chat_moderator_actions..456",
		},
	}

	for _, testCase := range testCases {
		c.Run(testCase.label, func(c *qt.C) {
			actual := ModerationActionTopic(testCase.inputUserID, testCase.inputChannelID)
			c.Assert(actual, qt.Equals, testCase.expected)
		})
	}
}

func TestParseModerationTopicChannelID(t *testing.T) {
	c := qt.New(t)

	type testCase struct {
		label             string
		inputTopic        string
		expectedChannelID string
		expectedErr       error
	}

	testCases := []testCase{
		{
			label:             "Standard",
			inputTopic:        "chat_moderator_actions.123.456",
			expectedChannelID: "456",
			expectedErr:       nil,
		},
		{
			label:             "Malformed",
			inputTopic:        "chat_moderator_actions.123",
			expectedChannelID: "",
			expectedErr:       errors.New("unable to parse channel ID from moderation topic"),
		},
	}

	for _, testCase := range testCases {
		c.Run(testCase.label, func(c *qt.C) {
			actualChannelID, err := parseChannelIDFromModerationTopic(testCase.inputTopic)
			if testCase.expectedErr == nil {
				c.Assert(err, qt.IsNil)
			} else {
				c.Assert(err, qt.ErrorMatches, testCase.expectedErr.Error())
			}
			c.Assert(actualChannelID, qt.Equals, testCase.expectedChannelID)
		})
	}
}
