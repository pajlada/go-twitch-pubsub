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
	if isModerationActionTopic(topic) {
		return messageTypeModerationAction
	}
	if isBitsEventTopic(topic) {
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
	if isSubscribeEventTopic(topic) {
		return messageTypeSubscribeEvent
	}

	return messageTypeUnknown
}
