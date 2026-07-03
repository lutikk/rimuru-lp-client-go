package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v3/api"
)

func MatchPing(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"пинг") {
			return true
		}
	}
	return false
}

func HandlePing(vk *api.VK, peerID, messageID, timestamp, ownerID, fromID int) {
	if fromID != ownerID {
		return
	}
	delta := time.Since(time.Unix(int64(timestamp), 0)).Seconds()
	if delta < 0 {
		delta = 666
	}
	text := fmt.Sprintf("ПОНГ Модуль ЛП\nОтвет через %.2f с", delta)
	EditMessage(vk, peerID, messageID, text)
}
