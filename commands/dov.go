package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/SevereCloud/vksdk/v3/api"
)

var userMentionRe = regexp.MustCompile(`(?:vk\.com/(?P<user>[\w.]+))|(?:\[id(?P<user_id>[\d]+)\|)`)

func MatchAddDov(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"+дов") {
			return true
		}
	}
	return false
}

func MatchRemoveDov(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"-дов") {
			return true
		}
	}
	return false
}

func MatchListDov(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if lower == p+"довы" {
			return true
		}
	}
	return false
}

func MatchDovPref(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"дов ") && !strings.HasPrefix(lower, p+"довы") {
			return true
		}
	}
	return false
}

func HandleAddDov(vk *api.VK, peerID, messageID, convMsgID, ownerID, fromID int, dovList []int, saveFn func([]int)) {
	if fromID != ownerID {
		return
	}

	userIDs := extractUserIDs(vk, peerID, convMsgID)
	if len(userIDs) == 0 {
		EditMessage(vk, peerID, messageID, "⚠ Не найден пользователь")
		return
	}

	for _, id := range dovList {
		if id == userIDs[0] {
			EditMessage(vk, peerID, messageID, "✅ Пользователь уже есть в доверенных")
			return
		}
	}

	saveFn(append(dovList, userIDs[0]))
	EditMessage(vk, peerID, messageID, "✅ Пользователь добавлен в доверенные ")
}

func HandleRemoveDov(vk *api.VK, peerID, messageID, convMsgID, ownerID, fromID int, dovList []int, saveFn func([]int)) {
	if fromID != ownerID {
		return
	}

	userIDs := extractUserIDs(vk, peerID, convMsgID)
	if len(userIDs) == 0 {
		EditMessage(vk, peerID, messageID, "⚠ Не найден пользователь")
		return
	}

	for i, id := range dovList {
		if id == userIDs[0] {
			newDov := append(dovList[:i], dovList[i+1:]...)
			saveFn(newDov)
			EditMessage(vk, peerID, messageID, "✅ Пользователь удален из игнора ")
			return
		}
	}
}

func HandleListDov(vk *api.VK, peerID, messageID, ownerID, fromID int, dovList []int) {
	if fromID != ownerID {
		return
	}

	text := "Вы доверяете:\n"
	if len(dovList) > 0 {
		ids := make([]string, len(dovList))
		for i, id := range dovList {
			ids[i] = strconv.Itoa(id)
		}
		resp, err := vk.UsersGet(api.Params{"user_ids": strings.Join(ids, ",")})
		if err == nil {
			for _, u := range resp {
				text += fmt.Sprintf("[id%d|%s %s]\n", u.ID, u.FirstName, u.LastName)
			}
		}
	}
	EditMessage(vk, peerID, messageID, text)
}

func HandleDovPref(vk *api.VK, peerID, messageID, ownerID, fromID int, text string, servPref []string, saveFn func(string)) {
	if fromID != ownerID {
		return
	}

	prefix := getServPrefix(servPref, text)
	newPref := strings.TrimSpace(text[len(prefix)+len("дов "):])
	saveFn(newPref)
	EditMessage(vk, peerID, messageID, fmt.Sprintf("Префикс изменен на <<%s>>", newPref))
}

func extractUserIDs(vk *api.VK, peerID, convMsgID int) []int {
	msg, err := GetMessageByConvID(vk, peerID, convMsgID)
	if err != nil {
		return nil
	}
	return SearchUserIDsFromMsg(vk, msg)
}

func SearchUserIDsFromMsg(vk *api.VK, msg map[string]interface{}) []int {
	var result []int
	seen := make(map[int]bool)

	addUnique := func(id int) {
		if id > 0 && !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}

	if text, ok := msg["text"].(string); ok && text != "" {
		matches := userMentionRe.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			if m[1] != "" {
				if uid, err := resolveScreen(vk, m[1]); err == nil {
					addUnique(uid)
				}
			}
			if m[2] != "" {
				if uid, err := strconv.Atoi(m[2]); err == nil {
					addUnique(uid)
				}
			}
		}
	}

	if reply, ok := msg["reply_message"].(map[string]interface{}); ok {
		if fromID, ok := reply["from_id"].(float64); ok && int(fromID) > 0 {
			addUnique(int(fromID))
		}
	}

	if fwds, ok := msg["fwd_messages"].([]interface{}); ok {
		for _, fwd := range fwds {
			if fwdMsg, ok := fwd.(map[string]interface{}); ok {
				if fromID, ok := fwdMsg["from_id"].(float64); ok && int(fromID) > 0 {
					addUnique(int(fromID))
				}
			}
		}
	}

	return result
}

func resolveScreen(vk *api.VK, domain string) (int, error) {
	raw, err := vk.Request("utils.resolveScreenName", api.Params{"screen_name": domain})
	if err != nil {
		return 0, err
	}
	var resp struct {
		Type     string `json:"type"`
		ObjectID int    `json:"object_id"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return 0, err
	}
	if resp.Type == "user" {
		return resp.ObjectID, nil
	}
	return 0, fmt.Errorf("not a user")
}
