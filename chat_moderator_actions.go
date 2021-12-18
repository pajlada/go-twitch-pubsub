package twitchpubsub

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMalformedChatModeratorActionsTopic = errors.New("malformed chat_moderator_actions topic")
)

/// INCOMING SECTION

// ChatModeratorAction describes an incoming "Moderation" action coming from Twitch's PubSub servers
type ChatModeratorAction struct {
	topic string

	Type             string   `json:"type"`
	ModerationAction string   `json:"moderation_action"`
	Arguments        []string `json:"args"`
	CreatedBy        string   `json:"created_by"`
	CreatedByUserID  string   `json:"created_by_user_id"`
	MsgID            string   `json:"msg_id"`
	TargetUserID     string   `json:"target_user_id"`
}

/// PARSER SECTION

type chatModeratorActionParser struct {
}

func (p *chatModeratorActionParser) Parse(topic string, innerMessageBytes []byte) (topicMessage, error) {
	type outerChatModeratorAction struct {
		Data ChatModeratorAction `json:"data"`
	}

	data := &outerChatModeratorAction{}
	err := json.Unmarshal(innerMessageBytes, data)
	if err != nil {
		return nil, fmt.Errorf("%w: chat_moderator_action inner payload bad: %s", ErrMalformedInnerPayload, err.Error())
	}

	data.Data.topic = topic

	return &data.Data, nil
}

func init() {
	registerParser("chat_moderator_actions", &chatModeratorActionParser{})
}

// parseChannelIDFromChatModeratorActionsTopic takes a topic and parses the channel ID from it
// Topic format: chat_moderator_actions.USERID.CHANNELID
func parseChannelIDFromChatModeratorActionsTopic(topic string) (string, error) {
	parts := strings.Split(topic, ".")
	if len(parts) != 3 {
		return "", ErrMalformedChatModeratorActionsTopic
	}

	return parts[2], nil
}

// handle is called from the client whenevera chat_moderator_actions message has been received
// this function is responsible for parsing out the relevant data from the topic and passing it along to the user
func (m *ChatModeratorAction) handle(c *Client) error {
	channelID, err := parseChannelIDFromChatModeratorActionsTopic(m.topic)
	if err != nil {
		return err
	}

	if c.onChatModerationAction == nil {
		return fmt.Errorf("%w: Use OnChatModeratorActions to specify your callback", ErrMissingCallback)
	}

	c.onChatModerationAction(channelID, m)

	return nil
}

/// LISTEN SECTION

// moderationActionTopic returns a properly formatted moderation action topic string with the given user and channel ID arguments
func createTopicChatModeratorActions(userID, channelID string) string {
	const f = `chat_moderator_actions.%s.%s`
	return fmt.Sprintf(f, userID, channelID)
}

// ListenChatModeratorActions listens to the chat_moderator_actions PubSub topic
// Required scope on authToken: channel:moderate
// Supports moderators listening to the topic, as well as users listening to the topic to receive their own events.
// Examples of moderator actions are bans, unbans, timeouts, deleting messages, changing chat mode (followers-only, subs-only), changing AutoMod levels, and adding a mod.
func (c *Client) ListenChatModeratorActions(userID, channelID, authToken string) {
	c.listen(createTopicChatModeratorActions(userID, channelID), authToken)
}

/// CALLBACK SECTION

// OnChatModeratorAction attaches the given callback to the moderation action event
func (c *Client) OnChatModeratorAction(callback func(channelID string, data *ChatModeratorAction)) {
	c.onChatModerationAction = callback
}
