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

	msg := &MessageEvent{
		MessageID:             msgID,
		Flags:                 flags,
		PeerID:                peerID,
		Timestamp:             timestamp,
		Text:                  text,
		FromID:                fromID,
		ConversationMessageID: convMsgID,
		Action:                action,
	}

	// User-longpoll НЕ отдаёт conversation_message_id в сыром массиве, а from_id
	// в личке приходится угадывать по peer_id (получается чужой id). Поэтому, как и
	// vkbottle (message_min -> messages.getById), догружаем полный объект сообщения
	// по локальному message_id и берём оттуда точные from_id/peer_id/text/cmid.
	// Без этого .л-сигнал уходит на хаб с conversation_message_id=0 (хаб «кладёт в
	// базу» вместо матча), а алиас-триггер бьёт getByConversationMessageId(0).
	// Сервисные (.слп) команды cmid не используют — поэтому и работали без гидрации.
	if action == "" && msgID != 0 {
		hydrateFromAPI(vk, msg)
	}

	return msg
}

// hydrateFromAPI догружает сообщение через messages.getById и перезаписывает
// поля точными значениями из VK. При ошибке оставляем то, что распарсили из LP.
func hydrateFromAPI(vk *api.VK, msg *MessageEvent) {
	resp, err := vk.MessagesGetByID(api.Params{"message_ids": msg.MessageID})
	if err != nil || len(resp.Items) == 0 {
		return
	}
	m := resp.Items[0]
	msg.FromID = m.FromID
	msg.PeerID = m.PeerID
	msg.Text = m.Text
	msg.ConversationMessageID = m.ConversationMessageID
	if m.Date != 0 {
		msg.Timestamp = m.Date
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
