package twitchpubsub

import (
	"errors"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func TestParseBitsEvent(t *testing.T) {
	c := qt.New(t)

	type testCase struct {
		label            string
		input            string
		isValidMsg       bool
		expected         *BitsEvent
		expectedErr      error
		expectedOuterErr error
	}

	testCases := []testCase{
		{
			label:      "1 bit",
			input:      `{"type":"MESSAGE","data":{"topic":"channel-bits-events-v1.11148817","message":"{\"data\":{\"user_name\":\"bbaper\",\"channel_name\":\"pajlada\",\"user_id\":\"165495734\",\"channel_id\":\"11148817\",\"time\":\"2023-06-17T15:39:51.276888655Z\",\"chat_message\":\"Cheer1 one free bit sir\",\"bits_used\":1,\"total_bits_used\":5,\"context\":\"cheer\",\"badge_entitlement\":null,\"badge_tier_entitlement\":{\"Badge\":{\"new_version\":0,\"previous_version\":0},\"Emoticons\":null}},\"version\":\"1.0\",\"message_type\":\"bits_event\",\"message_id\":\"540ee281-2f64-5463-ae85-ca79a6126037\"}"}}`,
			isValidMsg: true,
			expected: &BitsEvent{
				UserName: "bbaper",
				UserID:   "165495734",

				ChannelName: "pajlada",
				ChannelID:   "11148817",

				Time: time.Date(2023, time.June, 17, 15, 39, 51, 276888655, time.UTC),

				ChatMessage: "Cheer1 one free bit sir",

				BitsUsed: 1,

				TotalBitsUsed: 5,

				Context: "cheer",
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},
		{
			label:      "100 bit",
			input:      `{"type":"MESSAGE","data":{"topic":"channel-bits-events-v1.11148817","message":"{\"data\":{\"user_name\":\"slurps\",\"channel_name\":\"pajlada\",\"user_id\":\"133077169\",\"channel_id\":\"11148817\",\"time\":\"2023-06-17T15:41:15.524786977Z\",\"chat_message\":\"Cheer100  no problemo FeelsDankMan\",\"bits_used\":100,\"total_bits_used\":250,\"context\":\"cheer\",\"badge_entitlement\":null,\"badge_tier_entitlement\":{\"Badge\":{\"new_version\":0,\"previous_version\":0},\"Emoticons\":null}},\"version\":\"1.0\",\"message_type\":\"bits_event\",\"message_id\":\"2e7a028f-52fe-5f64-9d49-d7e8f500ebba\"}"}}`,
			isValidMsg: true,
			expected: &BitsEvent{
				UserName: "slurps",
				UserID:   "133077169",

				ChannelName: "pajlada",
				ChannelID:   "11148817",

				Time: time.Date(2023, time.June, 17, 15, 41, 15, 524786977, time.UTC),

				ChatMessage: "Cheer100  no problemo FeelsDankMan",

				BitsUsed: 100,

				TotalBitsUsed: 250,

				Context: "cheer",
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},
		{
			label:      "slurps 1 bit",
			input:      `{"type":"MESSAGE","data":{"topic":"channel-bits-events-v1.11148817","message":"{\"data\":{\"user_name\":\"slurps\",\"channel_name\":\"pajlada\",\"user_id\":\"133077169\",\"channel_id\":\"11148817\",\"time\":\"2023-06-17T15:43:18.755059824Z\",\"chat_message\":\"Cheer1\",\"bits_used\":1,\"total_bits_used\":251,\"context\":\"cheer\",\"badge_entitlement\":null,\"badge_tier_entitlement\":{\"Badge\":{\"new_version\":0,\"previous_version\":0},\"Emoticons\":null}},\"version\":\"1.0\",\"message_type\":\"bits_event\",\"message_id\":\"76b786ac-faf0-5eb8-8067-6e1a2fb56420\"}"}}`,
			isValidMsg: true,
			expected: &BitsEvent{
				UserName: "slurps",
				UserID:   "133077169",

				ChannelName: "pajlada",
				ChannelID:   "11148817",

				Time: time.Date(2023, time.June, 17, 15, 43, 18, 755059824, time.UTC),

				ChatMessage: "Cheer1",

				BitsUsed: 1,

				TotalBitsUsed: 251,

				Context: "cheer",
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},

		{
			label:            "Invalid message JSON",
			input:            `{"type":"MESSAGE","data":{"topic":"channel-bits-events-v1.11148817","message":"{forsen}"}}`,
			isValidMsg:       true,
			expected:         nil,
			expectedErr:      errors.New("invalid character 'f' looking for beginning of object key string"),
			expectedOuterErr: nil,
		},
	}

	for _, testCase := range testCases {
		c.Run(testCase.label, func(c *qt.C) {
			outerMessage, err := parseOuterMessage([]byte(testCase.input))
			c.Assert(err, qt.Equals, testCase.expectedOuterErr)
			c.Assert(isBitsEventTopic(outerMessage.Data.Topic), qt.Equals, testCase.isValidMsg)

			if testCase.isValidMsg {
				// Only test parsing if we expect the input message to be an actual subscribe message
				innerMessageBytes := []byte(outerMessage.Data.Message)
				actual, err := parseBitsEvent([]byte(innerMessageBytes))

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

func TestCreateBitsTopic(t *testing.T) {
	c := qt.New(t)

	type testCase struct {
		label          string
		inputChannelID string
		expected       string
	}

	testCases := []testCase{
		{
			label:          "Standard",
			inputChannelID: "456",
			expected:       "channel-bits-events-v1.456",
		},
		{
			label:          "Bad 1",
			inputChannelID: "forsen",
			expected:       "channel-bits-events-v1.forsen",
		},
		{
			label:          "Bad 2",
			inputChannelID: "",
			expected:       "channel-bits-events-v1.",
		},
	}

	for _, testCase := range testCases {
		c.Run(testCase.label, func(c *qt.C) {
			actual := BitsEventTopic(testCase.inputChannelID)
			c.Assert(actual, qt.Equals, testCase.expected)
		})
	}
}

func TestParseBitsTopicChannelID(t *testing.T) {
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
			inputTopic:        "channel-bits-events-v1.456",
			expectedChannelID: "456",
			expectedErr:       nil,
		},
		{
			label:             "Malformed but successful",
			inputTopic:        "channel-bits-events-v1.",
			expectedChannelID: "",
			expectedErr:       nil,
		},
		{
			label:             "Malformed",
			inputTopic:        "channel-bits-events-v1",
			expectedChannelID: "",
			expectedErr:       errors.New("unable to parse channel ID from bits topic"),
		},
	}

	for _, testCase := range testCases {
		c.Run(testCase.label, func(c *qt.C) {
			actualChannelID, err := parseChannelIDFromBitsTopic(testCase.inputTopic)
			if testCase.expectedErr == nil {
				c.Assert(err, qt.IsNil)
			} else {
				c.Assert(err, qt.ErrorMatches, testCase.expectedErr.Error())
			}
			c.Assert(actualChannelID, qt.Equals, testCase.expectedChannelID)
		})
	}
}
