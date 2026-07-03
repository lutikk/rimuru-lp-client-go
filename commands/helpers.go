package commands

import (
	"github.com/SevereCloud/vksdk/v3/api"
)

func EditMessage(vk *api.VK, peerID, messageID int, text string) {
	vk.MessagesEdit(api.Params{
		"peer_id":               peerID,
		"message_id":            messageID,
		"message":               text,
		"keep_forward_messages": true,
		"keep_snippets":         true,
		"dont_parse_links":      false,
	})
}
