package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/SevereCloud/vksdk/v3/api"
)

type IgnoreUser struct {
	UserID int `json:"user_id"`
	PeerID int `json:"peer_id"`
}

func MatchAddIgnore(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"+игнор") && !strings.Contains(lower, "глоигнор") {
			return true
		}
	}
	return false
}

func MatchRemoveIgnore(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"-игнор") && !strings.Contains(lower, "глоигнор") {
			return true
		}
	}
	return false
}

func MatchAddGlobalIgnore(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"+глоигнор") {
			return true
		}
	}
	return false
}

func MatchRemoveGlobalIgnore(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"-глоигнор") {
			return true
		}
	}
	return false
}

func MatchIgnoreList(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"игнор лист") {
			return true
		}
	}
	return false
}

func MatchGlobalIgnoreList(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"глоигнор лист") {
			return true
		}
	}
	return false
}

func HandleAddIgnore(vk *api.VK, peerID, messageID, convMsgID, ownerID, fromID int, ignoreList []IgnoreUser, saveFn func([]IgnoreUser)) {
	if fromID != ownerID {
		return
	}

	userIDs := extractUserIDs(vk, peerID, convMsgID)
	if len(userIDs) == 0 {
		EditMessage(vk, peerID, messageID, "⚠ Не найден пользователь")
		return
	}

	for _, ig := range ignoreList {
		if ig.PeerID == peerID && ig.UserID == userIDs[0] {
			EditMessage(vk, peerID, messageID, "✅ Этот пользователь уже в игноре")
			return
		}
	}

	newIgnore := IgnoreUser{UserID: userIDs[0], PeerID: peerID}
	saveFn(append(ignoreList, newIgnore))
	EditMessage(vk, peerID, messageID, "✅ Пользователь добавлен в игнор ")
}

func HandleRemoveIgnore(vk *api.VK, peerID, messageID, convMsgID, ownerID, fromID int, ignoreList []IgnoreUser, saveFn func([]IgnoreUser)) {
	if fromID != ownerID {
		return
	}

	userIDs := extractUserIDs(vk, peerID, convMsgID)
	if len(userIDs) == 0 {
		EditMessage(vk, peerID, messageID, "⚠ Не найден пользователь")
		return
	}

	for i, ig := range ignoreList {
		if ig.PeerID == peerID && ig.UserID == userIDs[0] {
			newList := append(ignoreList[:i], ignoreList[i+1:]...)
			saveFn(newList)
			EditMessage(vk, peerID, messageID, "✅ Пользователь удален из игнора ")
			return
		}
	}
}

func HandleAddGlobalIgnore(vk *api.VK, peerID, messageID, convMsgID, ownerID, fromID int, globalIgnore []int, saveFn func([]int)) {
	if fromID != ownerID {
		return
	}

	userIDs := extractUserIDs(vk, peerID, convMsgID)
	if len(userIDs) == 0 {
		EditMessage(vk, peerID, messageID, "⚠ Не найден пользователь")
		return
	}

	for _, id := range globalIgnore {
		if id == userIDs[0] {
			EditMessage(vk, peerID, messageID, "✅ Пользователь уже в глоигноре")
			return
		}
	}

	saveFn(append(globalIgnore, userIDs[0]))
	EditMessage(vk, peerID, messageID, "✅ Пользователь добавлен в глоигнор ")
}

func HandleRemoveGlobalIgnore(vk *api.VK, peerID, messageID, convMsgID, ownerID, fromID int, globalIgnore []int, saveFn func([]int)) {
	if fromID != ownerID {
		return
	}

	userIDs := extractUserIDs(vk, peerID, convMsgID)
	if len(userIDs) == 0 {
		EditMessage(vk, peerID, messageID, "⚠ Не найден пользователь")
		return
	}

	for i, id := range globalIgnore {
		if id == userIDs[0] {
			newList := append(globalIgnore[:i], globalIgnore[i+1:]...)
			saveFn(newList)
			EditMessage(vk, peerID, messageID, "✅ Пользователь удален из глоигнора")
			return
		}
	}
}

func HandleIgnoreList(vk *api.VK, peerID, messageID, ownerID, fromID int, ignoreList []IgnoreUser) {
	if fromID != ownerID {
		return
	}

	text := "В этом чате вы игнорируете:\n"
	for _, ig := range ignoreList {
		if ig.PeerID == peerID {
			resp, err := vk.UsersGet(api.Params{"user_ids": strconv.Itoa(ig.UserID)})
			if err == nil && len(resp) > 0 {
				text += fmt.Sprintf("[id%d|%s %s]\n", resp[0].ID, resp[0].FirstName, resp[0].LastName)
			}
		}
	}
	EditMessage(vk, peerID, messageID, text)
}

func HandleGlobalIgnoreList(vk *api.VK, peerID, messageID, ownerID, fromID int, globalIgnore []int) {
	if fromID != ownerID {
		return
	}

	text := "Вы игнорируете:\n"
	if len(globalIgnore) > 0 {
		ids := make([]string, len(globalIgnore))
		for i, id := range globalIgnore {
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
