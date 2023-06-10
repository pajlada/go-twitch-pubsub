package twitchpubsub

import "strings"

// Base TODO: Refactor
type Base struct {
	Type string `json:"type"`
}

type messageType = int

const (
	messageTypeUnknown messageType = iota
	messageTypeModerationAction
	messageTypeBitsEvent
	messageTypePointsEvent
	messageTypeAutoModQueueEvent
	messageTypeWhisperEvent
	messageTypeSubscribeEvent
)

func getMessageType(topic string) messageType {
	if strings.HasPrefix(topic, moderationActionTopicPrefix) {
		return messageTypeModerationAction
	}
	if strings.HasPrefix(topic, bitsEventTopicPrefix) {
		return messageTypeBitsEvent
	}
	if strings.HasPrefix(topic, pointsEventTopicPrefix) {
		return messageTypePointsEvent
	}
	if strings.HasPrefix(topic, autoModQueueEventTopicPrefix) {
		return messageTypeAutoModQueueEvent
	}
	if strings.HasPrefix(topic, whisperEventTopicPrefix) {
		return messageTypeWhisperEvent
	}
	if strings.HasPrefix(topic, subscribeEventTopicPrefix) {
		return messageTypeSubscribeEvent
	}

	return messageTypeUnknown
}
