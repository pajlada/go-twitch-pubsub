package twitchpubsub

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func TestParseSubscribeEvent(t *testing.T) {
	c := qt.New(t)

	type testCase struct {
		label            string
		input            string
		isValidMsg       bool
		expected         *SubscribeEvent
		expectedErr      error
		expectedOuterErr error
	}

	testCases := []testCase{
		{
			label:      "Prime resubscription with no emotes & no message",
			input:      `{"type":"MESSAGE","data":{"topic":"channel-subscribe-events-v1.11148817","message":"{\"benefit_end_month\":0,\"user_name\":\"randers\",\"display_name\":\"randers\",\"channel_name\":\"pajlada\",\"user_id\":\"40286300\",\"channel_id\":\"11148817\",\"time\":\"2023-06-11T10:44:06.975336457Z\",\"sub_message\":{\"message\":\"\",\"emotes\":null},\"sub_plan\":\"Prime\",\"sub_plan_name\":\"look at those shitty emotes, rip $5 LUL\",\"months\":0,\"cumulative_months\":54,\"context\":\"resub\",\"is_gift\":false,\"multi_month_duration\":0}"}}`,
			isValidMsg: true,
			expected: &SubscribeEvent{
				ChannelID:            "11148817",
				ChannelName:          "pajlada",
				UserID:               "40286300",
				UserName:             "randers",
				DisplayName:          "randers",
				RecipientID:          "",
				RecipientUserName:    "",
				RecipientDisplayName: "",
				Time:                 time.Date(2023, time.June, 11, 10, 44, 6, 975336457, time.UTC),
				SubPlan:              "Prime",
				SubPlanName:          "look at those shitty emotes, rip $5 LUL",
				CumulativeMonths:     54,
				StreakMonths:         0,
				Context:              "resub",
				IsGift:               false,
				SubMessage: SubMessage{
					Message: "",
					Emotes:  nil,
				},
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},
		{
			label:      "Prime resubscription with no emotes",
			input:      `{"type":"MESSAGE","data":{"topic":"channel-subscribe-events-v1.11148817","message":"{\"benefit_end_month\":0,\"user_name\":\"supersaintnick\",\"display_name\":\"SuperSaintNick\",\"channel_name\":\"pajlada\",\"user_id\":\"123747906\",\"channel_id\":\"11148817\",\"time\":\"2023-06-11T11:39:00.678953302Z\",\"sub_message\":{\"message\":\"pajaCheese\",\"emotes\":null},\"sub_plan\":\"Prime\",\"sub_plan_name\":\"look at those shitty emotes, rip $5 LUL\",\"months\":0,\"cumulative_months\":2,\"context\":\"resub\",\"is_gift\":false,\"multi_month_duration\":0}"}}`,
			isValidMsg: true,
			expected: &SubscribeEvent{
				ChannelID:            "11148817",
				ChannelName:          "pajlada",
				UserID:               "123747906",
				UserName:             "supersaintnick",
				DisplayName:          "SuperSaintNick",
				RecipientID:          "",
				RecipientUserName:    "",
				RecipientDisplayName: "",
				Time:                 time.Date(2023, time.June, 11, 11, 39, 0, 678953302, time.UTC),
				SubPlan:              "Prime",
				SubPlanName:          "look at those shitty emotes, rip $5 LUL",
				CumulativeMonths:     2,
				StreakMonths:         0,
				Context:              "resub",
				IsGift:               false,
				SubMessage: SubMessage{
					Message: "pajaCheese",
					Emotes:  nil,
				},
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},
		{
			label:      "Non-prime resubscription with emotes",
			input:      `{"type":"MESSAGE","data":{"topic":"channel-subscribe-events-v1.11148817","message":"{\"benefit_end_month\":0,\"user_name\":\"nerixyz\",\"display_name\":\"nerixyz\",\"channel_name\":\"pajlada\",\"user_id\":\"129546453\",\"channel_id\":\"11148817\",\"time\":\"2023-06-11T11:22:45.552399915Z\",\"sub_message\":{\"message\":\"forsen pajaW FeelsDankMan m0xyGift\",\"emotes\":[{\"start\":7,\"end\":11,\"id\":\"80481\"},{\"start\":26,\"end\":33,\"id\":\"303660952\"}]},\"sub_plan\":\"1000\",\"sub_plan_name\":\"look at those shitty emotes, rip $5 LUL\",\"months\":0,\"cumulative_months\":12,\"context\":\"resub\",\"is_gift\":false,\"multi_month_duration\":0}"}}`,
			isValidMsg: true,
			expected: &SubscribeEvent{
				ChannelID:            "11148817",
				ChannelName:          "pajlada",
				UserID:               "129546453",
				UserName:             "nerixyz",
				DisplayName:          "nerixyz",
				RecipientID:          "",
				RecipientUserName:    "",
				RecipientDisplayName: "",
				Time:                 time.Date(2023, time.June, 11, 11, 22, 45, 552399915, time.UTC),
				SubPlan:              "1000",
				SubPlanName:          "look at those shitty emotes, rip $5 LUL",
				CumulativeMonths:     12,
				StreakMonths:         0,
				Context:              "resub",
				IsGift:               false,
				SubMessage: SubMessage{
					Message: "forsen pajaW FeelsDankMan m0xyGift",
					Emotes: []Emotes{
						{
							Start: 7,
							End:   11,
							ID:    "80481",
						},
						{
							Start: 26,
							End:   33,
							ID:    "303660952",
						},
					},
				},
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},
		{
			label:      "Non-prime resubscription with streak",
			input:      `{"type":"MESSAGE","data":{"topic":"channel-subscribe-events-v1.11148817","message":"{\"benefit_end_month\":0,\"user_name\":\"rotten_pizza\",\"display_name\":\"rotten_pizza\",\"channel_name\":\"pajlada\",\"user_id\":\"69478316\",\"channel_id\":\"11148817\",\"time\":\"2023-06-11T11:33:01.381805678Z\",\"sub_message\":{\"message\":\"!#playsound cheese pajaCheese\",\"emotes\":null},\"sub_plan\":\"1000\",\"sub_plan_name\":\"look at those shitty emotes, rip $5 LUL\",\"months\":0,\"cumulative_months\":24,\"streak_months\":24,\"context\":\"resub\",\"is_gift\":false,\"multi_month_duration\":0}"}}`,
			isValidMsg: true,
			expected: &SubscribeEvent{
				ChannelID:            "11148817",
				ChannelName:          "pajlada",
				UserID:               "69478316",
				UserName:             "rotten_pizza",
				DisplayName:          "rotten_pizza",
				RecipientID:          "",
				RecipientUserName:    "",
				RecipientDisplayName: "",
				Time:                 time.Date(2023, time.June, 11, 11, 33, 01, 381805678, time.UTC),
				SubPlan:              "1000",
				SubPlanName:          "look at those shitty emotes, rip $5 LUL",
				CumulativeMonths:     24,
				StreakMonths:         24,
				Context:              "resub",
				IsGift:               false,
				SubMessage: SubMessage{
					Message: "!#playsound cheese pajaCheese",
					Emotes:  nil,
				},
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},
		{
			label:      "Gift subscription",
			input:      `{"type":"MESSAGE","data":{"topic":"channel-subscribe-events-v1.11148817","message":"{\"benefit_end_month\":0,\"user_name\":\"pajlada\",\"display_name\":\"pajlada\",\"channel_name\":\"pajlada\",\"user_id\":\"11148817\",\"channel_id\":\"11148817\",\"recipient_id\":\"29294552\",\"recipient_user_name\":\"ultrainstinctcocabear\",\"recipient_display_name\":\"ultrainstinctcocabear\",\"time\":\"2023-06-11T11:36:21.198150864Z\",\"sub_message\":{\"message\":\"\",\"emotes\":null},\"sub_plan\":\"1000\",\"sub_plan_name\":\"look at those shitty emotes, rip $5 LUL\",\"months\":2,\"context\":\"subgift\",\"is_gift\":true,\"multi_month_duration\":1}"}}`,
			isValidMsg: true,
			expected: &SubscribeEvent{
				ChannelID:            "11148817",
				ChannelName:          "pajlada",
				UserID:               "11148817", // Sender ID
				UserName:             "pajlada",  // Sender login
				DisplayName:          "pajlada",  // Sender display name
				RecipientID:          "29294552",
				RecipientUserName:    "ultrainstinctcocabear",
				RecipientDisplayName: "ultrainstinctcocabear",
				Time:                 time.Date(2023, time.June, 11, 11, 36, 21, 198150864, time.UTC),
				SubPlan:              "1000",
				SubPlanName:          "look at those shitty emotes, rip $5 LUL",
				CumulativeMonths:     0,
				StreakMonths:         0,
				Context:              "subgift",
				IsGift:               true,
				SubMessage: SubMessage{
					Message: "",
					Emotes:  nil,
				},
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},
		{
			label:      "Anonymous gift subscription",
			input:      `{"type":"MESSAGE","data":{"topic":"channel-subscribe-events-v1.11148817","message":"{\"benefit_end_month\":0,\"channel_name\":\"pajlada\",\"channel_id\":\"11148817\",\"recipient_id\":\"72902587\",\"recipient_user_name\":\"slch000\",\"recipient_display_name\":\"SLCH000\",\"time\":\"2023-06-11T11:43:08.367893236Z\",\"sub_message\":{\"message\":\"\",\"emotes\":null},\"sub_plan\":\"1000\",\"sub_plan_name\":\"look at those shitty emotes, rip $5 LUL\",\"months\":3,\"context\":\"anonsubgift\",\"is_gift\":true,\"multi_month_duration\":1}"}}`,
			isValidMsg: true,
			expected: &SubscribeEvent{
				ChannelID:            "11148817",
				ChannelName:          "pajlada",
				UserID:               "", // Sender ID
				UserName:             "", // Sender login
				DisplayName:          "", // Sender display name
				RecipientID:          "72902587",
				RecipientUserName:    "slch000",
				RecipientDisplayName: "SLCH000",
				Time:                 time.Date(2023, time.June, 11, 11, 43, 8, 367893236, time.UTC),
				SubPlan:              "1000",
				SubPlanName:          "look at those shitty emotes, rip $5 LUL",
				CumulativeMonths:     0,
				StreakMonths:         0,
				Context:              "anonsubgift",
				IsGift:               true,
				SubMessage: SubMessage{
					Message: "",
					Emotes:  nil,
				},
			},
			expectedErr:      nil,
			expectedOuterErr: nil,
		},
	}

	for _, testCase := range testCases {
		c.Run(testCase.label, func(c *qt.C) {
			outerMessage, err := parseOuterMessage([]byte(testCase.input))
			c.Assert(err, qt.Equals, testCase.expectedOuterErr)
			c.Assert(isSubscribeEventTopic(outerMessage.Data.Topic), qt.Equals, testCase.isValidMsg)

			if testCase.isValidMsg {
				// Only test parsing if we expect the input message to be an actual subscribe message
				innerMessageBytes := []byte(outerMessage.Data.Message)
				actual, err := parseSubscribeEvent([]byte(innerMessageBytes))

				c.Assert(err, qt.Equals, testCase.expectedErr)
				c.Assert(actual, qt.DeepEquals, testCase.expected)
			}
		})
	}
}
