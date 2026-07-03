package commands

import (
	"fmt"
	"strings"

	"github.com/SevereCloud/vksdk/v3/api"
)

type Alias struct {
	Name    string `json:"name"`
	SokrCmd string `json:"sokr_cmd"`
	PolnCmd string `json:"poln_cmd"`
}

type AliasStore interface {
	GetAliases() []Alias
	AddAlias(a Alias)
	RemoveAlias(name string) bool
	Save() error
}

func MatchAddAlias(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"+алиас ") {
			return true
		}
	}
	return false
}

func MatchRemoveAlias(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"-алиас ") {
			return true
		}
	}
	return false
}

func MatchListAliases(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"алиасы") {
			return true
		}
	}
	return false
}

func HandleAddAlias(vk *api.VK, peerID, messageID, ownerID, fromID int, text string, servPref []string, aliases []Alias, saveFn func([]Alias)) {
	if fromID != ownerID {
		return
	}

	prefix := getServPrefix(servPref, text)
	rest := strings.TrimPrefix(strings.ToLower(text), prefix+"+алиас ")
	rest = text[len(prefix)+len("+алиас "):]

	lines := strings.SplitN(rest, "\n", 3)
	if len(lines) < 3 {
		EditMessage(vk, peerID, messageID, "⚠ Формат: +алиас <имя>\\n<сокращение>\\n<команда>")
		return
	}

	name := strings.TrimSpace(lines[0])
	sokrCmd := strings.TrimSpace(lines[1])
	polnCmd := strings.TrimSpace(lines[2])

	for _, a := range aliases {
		if a.Name == name {
			EditMessage(vk, peerID, messageID, "Такой алиас уже существует")
			return
		}
	}

	newAlias := Alias{Name: name, SokrCmd: sokrCmd, PolnCmd: polnCmd}
	saveFn(append(aliases, newAlias))
	EditMessage(vk, peerID, messageID, fmt.Sprintf("Создал алиас <<%s>>", name))
}

func HandleRemoveAlias(vk *api.VK, peerID, messageID, ownerID, fromID int, text string, servPref []string, aliases []Alias, saveFn func([]Alias)) {
	if fromID != ownerID {
		return
	}

	prefix := getServPrefix(servPref, text)
	name := strings.TrimSpace(text[len(prefix)+len("-алиас "):])

	for i, a := range aliases {
		if a.Name == name {
			newAliases := append(aliases[:i], aliases[i+1:]...)
			saveFn(newAliases)
			EditMessage(vk, peerID, messageID, fmt.Sprintf("Алиас <<%s>> удален", name))
			return
		}
	}
	EditMessage(vk, peerID, messageID, fmt.Sprintf("Алиас <<%s>> не найден", name))
}

func HandleListAliases(vk *api.VK, peerID, messageID, ownerID, fromID int, aliases []Alias) {
	if fromID != ownerID {
		return
	}

	text := "📃 Ваши алиасы:\n"
	for i, a := range aliases {
		text += fmt.Sprintf("%d. %s (%s -> .л %s)\n", i+1, a.Name, a.SokrCmd, a.PolnCmd)
	}
	EditMessage(vk, peerID, messageID, text)
}

func getServPrefix(servPref []string, text string) string {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p) {
			return p
		}
	}
	return ""
}
