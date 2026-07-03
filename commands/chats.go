package commands

import (
	"strings"

	"github.com/SevereCloud/vksdk/v3/api"
)

type HelloChat struct {
	PeerID    int    `json:"peer_id"`
	HelloText string `json:"hello_text"`
}

func MatchAddHello(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"+приветствие") {
			return true
		}
	}
	return false
}

func MatchRemoveHello(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"-приветствие") {
			return true
		}
	}
	return false
}

func HandleChatInvite(vk *api.VK, peerID int, helloList []HelloChat) {
	for _, h := range helloList {
		if h.PeerID == peerID {
			vk.MessagesSend(api.Params{
				"peer_id":   peerID,
				"message":   h.HelloText,
				"random_id": 0,
			})
			return
		}
	}
}

func HandleAddHello(vk *api.VK, peerID, messageID, ownerID, fromID int, text string, servPref []string, helloList []HelloChat, saveFn func([]HelloChat)) {
	if fromID != ownerID {
		return
	}

	prefix := getServPrefix(servPref, text)
	rest := text[len(prefix)+len("+приветствие"):]
	helloText := strings.TrimLeft(rest, " \n")

	if helloText == "" {
		EditMessage(vk, peerID, messageID, "⚠ Укажите текст приветствия")
		return
	}

	for _, h := range helloList {
		if h.PeerID == peerID {
			EditMessage(vk, peerID, messageID, "В этом чате уже установлено приветствие")
			return
		}
	}

	newHello := HelloChat{PeerID: peerID, HelloText: helloText}
	saveFn(append(helloList, newHello))
	EditMessage(vk, peerID, messageID, "Добавлено приветствие в этот чат")
}

func HandleRemoveHello(vk *api.VK, peerID, messageID, ownerID, fromID int, helloList []HelloChat, saveFn func([]HelloChat)) {
	if fromID != ownerID {
		return
	}

	for i, h := range helloList {
		if h.PeerID == peerID {
			newList := append(helloList[:i], helloList[i+1:]...)
			saveFn(newList)
			EditMessage(vk, peerID, messageID, "Удалил приветствие")
			return
		}
	}

	EditMessage(vk, peerID, messageID, "Приветствие не найдено")
}
