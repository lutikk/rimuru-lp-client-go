package commands

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/SevereCloud/vksdk/v3/api"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func HandleSendSignal(vk *api.VK, peerID, messageID, convMsgID, fromID, ownerID, timestamp int, text, secretCode, callbackURL string) {
	if fromID != ownerID {
		return
	}

	model := map[string]interface{}{
		"user_id": fromID,
		"method":  "lpSendMySignal",
		"secret":  secretCode,
		"message": map[string]interface{}{
			"conversation_message_id": convMsgID,
			"from_id":                 fromID,
			"date":                    timestamp,
			"text":                    ".с " + text,
			"peer_id":                 peerID,
		},
		"object": map[string]interface{}{
			"chat":                    nil,
			"from_id":                 fromID,
			"value":                   text,
			"conversation_message_id": convMsgID,
		},
	}

	sendCallback(vk, model, peerID, messageID, callbackURL)
}

func HandleAliasSignal(vk *api.VK, peerID, messageID, convMsgID, fromID, ownerID, timestamp int, polnCmd, signal, separator, secretCode, callbackURL, token string) {
	if fromID != ownerID {
		return
	}

	msg, err := GetMessageByConvID(vk, peerID, convMsgID)
	if err != nil {
		EditMessage(vk, peerID, messageID, fmt.Sprintf("⚠ Ошибка: %v", err))
		return
	}

	msgFromID := jsonInt(msg, "from_id")
	msgConvID := jsonInt(msg, "conversation_message_id")

	preparedText := polnCmd
	if signal != "" {
		preparedText += separator + signal
	}

	model := map[string]interface{}{
		"user_id": msgFromID,
		"method":  "lpSendMySignal",
		"secret":  secretCode,
		"message": map[string]interface{}{
			"conversation_message_id": msgConvID,
			"from_id":                 msgFromID,
			"date":                    timestamp,
			"text":                    preparedText,
			"peer_id":                 peerID,
		},
		"object": map[string]interface{}{
			"chat":                    nil,
			"from_id":                 msgFromID,
			"value":                   preparedText,
			"conversation_message_id": msgConvID,
		},
	}

	sendCallback(vk, model, peerID, messageID, callbackURL)
}

func GetMessageByConvID(vk *api.VK, peerID, convMsgID int) (map[string]interface{}, error) {
	raw, err := vk.Request("messages.getByConversationMessageId", api.Params{
		"peer_id":                  peerID,
		"conversation_message_ids": convMsgID,
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("сообщение не найдено")
	}
	return resp.Items[0], nil
}

type callbackResponse struct {
	Response     string `json:"response"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func sendCallback(vk *api.VK, data map[string]interface{}, peerID, messageID int, callbackURL string) {
	body, _ := json.Marshal(data)
	resp, err := httpClient.Post(callbackURL, "application/json", bytes.NewReader(body))

	var errMsg string

	if err != nil {
		errMsg = fmt.Sprintf("⚠ Ошибка сети: %v", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			errMsg = fmt.Sprintf("⚠ Ошибка сервера Rimulu. Сервер, ответил кодом %d.", resp.StatusCode)
		} else {
			respBody, _ := io.ReadAll(resp.Body)
			var cr callbackResponse
			json.Unmarshal(respBody, &cr)

			switch cr.Response {
			case "ok":
				return
			case "error":
				errMsg = formatErrorMsg(cr.ErrorCode)
			case "vk_error":
				errMsg = fmt.Sprintf("⚠ Ошибка сервера Rimulu. Сервер, ответил: <<Ошибка VK %d %s>>",
					cr.ErrorCode, cr.ErrorMessage)
			default:
				errMsg = fmt.Sprintf("⚠ Неизвестный ответ: %s", cr.Response)
			}
		}
	}

	if errMsg != "" {
		EditMessage(vk, peerID, messageID, errMsg)
	}
}

func formatErrorMsg(code int) string {
	messages := map[int]string{
		1:  "Пустой запрос",
		2:  "Неизвестный тип сигнала",
		3:  "Пара пользователь/секрет не найдены",
		4:  "Беседа не привязана",
		10: "Не удалось связать беседу",
	}
	if msg, ok := messages[code]; ok {
		return fmt.Sprintf("⚠ Ошибка сервера Rimulu. Сервер, ответил: <<%s>>", msg)
	}
	return fmt.Sprintf("⚠ Ошибка сервера Rimulu. Сервер, ответил: <<Ошибка #%d>>", code)
}

func jsonInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}
