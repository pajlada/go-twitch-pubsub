package twitchpubsub

// Helper functions and structures for twitch bits events

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const pointsEventTopicPrefix = "channel-points-channel-v1."

// PointsEvent describes an incoming "Channel Points" action coming from Twitch's PubSub servers
type PointsEvent struct {
	Id   string `json:"id"`
	User struct {
		Id          string `json:"id"`
		User        string `json:"login"`
		DisplayName string `json:"display_name"`
	} `json:""user`
	ChannelID  string    `json:"channel_id"`
	RedeemedAt time.Time `json:"redeemed_at"`
	Reward     struct {
		Id           string `json:"id"`
		Title        string `json:"title"`
		Desc         string `json:"prompt"`
		Cost         int    `json:"cost"`
		UserInputReq bool   `json:"is_user_input_required"`
		SubOnly      bool   `json:"is_sub_only"`
		Enabled      bool   `json:"is_enabled"`
		Paused       bool   `json:"is_paused"`
		InStock      bool   `json:"is_in_stock"`
		MaxPerStream struct {
			Enabled bool `json:"is_enabled"`
			Max     int  `json:"max_per_stream"`
		} `json:"max_per_stream"`
		MaxPerUserPerStream struct {
			Enabled bool `json:"is_enabled"`
			Max     int  `json:"max_per_user_per_stream"`
		} `json:"max_per_user_per_stream"`
		GlobalCooldown struct {
			Enabled bool `json:"is_enabled"`
			Cd      int  `json:"global_cooldown_seconds"`
		} `json:"global_cooldown"`
	} `json:"reward"`
}

type outerPointsEvent struct {
	Data pointsEventData `json:"data"`
}

type pointsEventData struct {
	Redemption PointsEvent `json:"redemption"`
}

func parsePointsEvent(bytes []byte) (*PointsEvent, error) {
	data := &outerPointsEvent{}
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return nil, err
	}
	return &data.Data.Redemption, nil
}

func parseChannelIDFromPointsTopic(topic string) (string, error) {
	parts := strings.Split(topic, ".")
	if len(parts) != 2 {
		return "", errors.New("Unable to parse channel ID from points topic")
	}

	return parts[1], nil
}

// PointsEventTopic returns a properly formatted points event topic string with the given channel ID argument
func PointsEventTopic(channelID string) string {
	const f = `channel-points-channel-v1.%s`
	return fmt.Sprintf(f, channelID)
}
