package main

import (
	"strconv"

	"github.com/SevereCloud/vksdk/v3/api"
)

type MessageEvent struct {
	MessageID             int
	Flags                 int
	PeerID                int
	Timestamp             int
	Text                  string
	FromID                int
	ConversationMessageID int
	Action                string
}

func parseMessageEvent(event []interface{}, vk *api.VK) *MessageEvent {
	if len(event) < 6 {
		return nil
	}

	msgID := toInt(event[1])
	flags := toInt(event[2])
	peerID := toInt(event[3])
	timestamp := toInt(event[4])
	text := toString(event[5])

	var fromID int
	var convMsgID int
	var action string

	if len(event) > 6 {
		if extra, ok := event[6].(map[string]interface{}); ok {
			if from, ok := extra["from"].(string); ok {
				fromID, _ = strconv.Atoi(from)
			}
			if sa, ok := extra["source_act"].(string); ok {
				action = sa
			}
		}
	}

	if fromID == 0 {
		if peerID < 2000000000 {
			fromID = peerID
		}
	}

	if len(event) > 7 {
		if extra2, ok := event[7].(map[string]interface{}); ok {
			if cmid, ok := extra2["conversation_message_id"].(string); ok {
				convMsgID, _ = strconv.Atoi(cmid)
			} else if cmid, ok := extra2["conversation_message_id"].(float64); ok {
				convMsgID = int(cmid)
			}
		}
	}

	return &MessageEvent{
		MessageID:             msgID,
		Flags:                 flags,
		PeerID:                peerID,
		Timestamp:             timestamp,
		Text:                  text,
		FromID:                fromID,
		ConversationMessageID: convMsgID,
		Action:                action,
	}
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		i, _ := strconv.Atoi(val)
		return i
	}
	return 0
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
