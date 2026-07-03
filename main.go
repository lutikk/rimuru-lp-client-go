package main

import (
	"os"
	"strings"

	"github.com/SevereCloud/vksdk/v3/api"
	longpoll "github.com/SevereCloud/vksdk/v3/longpoll-user"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"rimuru_lp_client_go/commands"
)

// version — версия сборки, проставляется линкером: -ldflags "-X main.version=vX.Y.Z".
var version = "v5.0.0"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Str("version", version).Msg("Rimuru LP client")

	cfg := LoadConfig()
	log.Info().Int("user_id", cfg.ID).Msg("Конфиг загружен")

	code, err := getCode(cfg.Token)
	if err != nil {
		log.Fatal().Err(err).Msg("Не удалось получить secret_code")
	}
	cfg.SecretCode = code
	cfg.Save()
	log.Info().Msg("Secret code получен")

	vk := api.NewVK(cfg.Token)

	mode := longpoll.ReceiveAttachments + longpoll.ExtendedEvents
	lp, err := longpoll.NewLongPoll(vk, mode)
	if err != nil {
		log.Fatal().Err(err).Msg("Не удалось создать Long Poll")
	}
	lp.Goroutine(true)

	lp.EventNew(4, func(event []interface{}) error {
		handleMessage(vk, cfg, event)
		return nil
	})

	lp.EventNew(4, func(event []interface{}) error {
		return nil
	})

	log.Info().Msg("RimuruLP (Go) запущен, слушаю события...")

	if err := lp.Run(); err != nil {
		log.Fatal().Err(err).Msg("Long Poll остановлен с ошибкой")
	}
}

func handleMessage(vk *api.VK, cfg *UserConfig, event []interface{}) {
	msg := parseMessageEvent(event, vk)
	if msg == nil {
		return
	}

	if msg.Action != "" {
		if msg.Action == "chat_invite_user" || msg.Action == "chat_invite_user_by_link" {
			handleChatInvite(vk, cfg, msg)
		}
		return
	}

	if msg.Text == "" {
		return
	}

	if !middleware(vk, cfg, msg) {
		return
	}

	text := msg.Text

	// --- Сервисные команды (по приоритету) ---

	if commands.MatchPing(cfg.ServPref, text) {
		commands.HandlePing(vk, msg.PeerID, msg.MessageID, msg.Timestamp, cfg.ID, msg.FromID)
		return
	}

	if commands.MatchInfo(cfg.ServPref, text) {
		commands.HandleInfo(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, commands.InfoData{
			MyPref:       cfg.MyPref,
			ServPref:     cfg.ServPref,
			AliasCount:   len(cfg.Aliases),
			DovCount:     len(cfg.Dov),
			DovPref:      cfg.DovPref,
			IgnoreCount:  len(cfg.Ignore),
			GIgnoreCount: len(cfg.GlobalIgnore),
			HelloCount:   len(cfg.Hello),
		}, versionLink())
		return
	}

	if commands.MatchServer(cfg.ServPref, text) {
		commands.HandleServer(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID)
		return
	}

	if commands.MatchAddPref(cfg.ServPref, text) {
		commands.HandleAddPref(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, text, cfg.ServPref, cfg.MyPref, func(newPref []string) {
			cfg.MyPref = newPref
			cfg.Save()
		})
		return
	}

	if commands.MatchRemovePref(cfg.ServPref, text) {
		commands.HandleRemovePref(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, text, cfg.ServPref, cfg.MyPref, func(newPref []string) {
			cfg.MyPref = newPref
			cfg.Save()
		})
		return
	}

	if commands.MatchAddAlias(cfg.ServPref, text) {
		aliases := toCommandAliases(cfg.Aliases)
		commands.HandleAddAlias(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, text, cfg.ServPref, aliases, func(newAliases []commands.Alias) {
			cfg.Aliases = fromCommandAliases(newAliases)
			cfg.Save()
		})
		return
	}

	if commands.MatchRemoveAlias(cfg.ServPref, text) {
		aliases := toCommandAliases(cfg.Aliases)
		commands.HandleRemoveAlias(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, text, cfg.ServPref, aliases, func(newAliases []commands.Alias) {
			cfg.Aliases = fromCommandAliases(newAliases)
			cfg.Save()
		})
		return
	}

	if commands.MatchListAliases(cfg.ServPref, text) {
		commands.HandleListAliases(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, toCommandAliases(cfg.Aliases))
		return
	}

	if commands.MatchAddDov(cfg.ServPref, text) {
		commands.HandleAddDov(vk, msg.PeerID, msg.MessageID, msg.ConversationMessageID, cfg.ID, msg.FromID, cfg.Dov, func(newDov []int) {
			cfg.Dov = newDov
			cfg.Save()
		})
		return
	}

	if commands.MatchRemoveDov(cfg.ServPref, text) {
		commands.HandleRemoveDov(vk, msg.PeerID, msg.MessageID, msg.ConversationMessageID, cfg.ID, msg.FromID, cfg.Dov, func(newDov []int) {
			cfg.Dov = newDov
			cfg.Save()
		})
		return
	}

	if commands.MatchListDov(cfg.ServPref, text) {
		commands.HandleListDov(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, cfg.Dov)
		return
	}

	if commands.MatchDovPref(cfg.ServPref, text) {
		commands.HandleDovPref(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, text, cfg.ServPref, func(newPref string) {
			cfg.DovPref = newPref
			cfg.Save()
		})
		return
	}

	if commands.MatchGlobalIgnoreList(cfg.ServPref, text) {
		commands.HandleGlobalIgnoreList(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, cfg.GlobalIgnore)
		return
	}

	if commands.MatchIgnoreList(cfg.ServPref, text) {
		ignoreList := toCommandIgnoreUsers(cfg.Ignore)
		commands.HandleIgnoreList(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, ignoreList)
		return
	}

	if commands.MatchAddGlobalIgnore(cfg.ServPref, text) {
		commands.HandleAddGlobalIgnore(vk, msg.PeerID, msg.MessageID, msg.ConversationMessageID, cfg.ID, msg.FromID, cfg.GlobalIgnore, func(newList []int) {
			cfg.GlobalIgnore = newList
			cfg.Save()
		})
		return
	}

	if commands.MatchRemoveGlobalIgnore(cfg.ServPref, text) {
		commands.HandleRemoveGlobalIgnore(vk, msg.PeerID, msg.MessageID, msg.ConversationMessageID, cfg.ID, msg.FromID, cfg.GlobalIgnore, func(newList []int) {
			cfg.GlobalIgnore = newList
			cfg.Save()
		})
		return
	}

	if commands.MatchAddIgnore(cfg.ServPref, text) {
		ignoreList := toCommandIgnoreUsers(cfg.Ignore)
		commands.HandleAddIgnore(vk, msg.PeerID, msg.MessageID, msg.ConversationMessageID, cfg.ID, msg.FromID, ignoreList, func(newList []commands.IgnoreUser) {
			cfg.Ignore = fromCommandIgnoreUsers(newList)
			cfg.Save()
		})
		return
	}

	if commands.MatchRemoveIgnore(cfg.ServPref, text) {
		ignoreList := toCommandIgnoreUsers(cfg.Ignore)
		commands.HandleRemoveIgnore(vk, msg.PeerID, msg.MessageID, msg.ConversationMessageID, cfg.ID, msg.FromID, ignoreList, func(newList []commands.IgnoreUser) {
			cfg.Ignore = fromCommandIgnoreUsers(newList)
			cfg.Save()
		})
		return
	}

	if commands.MatchAddHello(cfg.ServPref, text) {
		helloList := toCommandHelloChats(cfg.Hello)
		commands.HandleAddHello(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, text, cfg.ServPref, helloList, func(newList []commands.HelloChat) {
			cfg.Hello = fromCommandHelloChats(newList)
			cfg.Save()
		})
		return
	}

	if commands.MatchRemoveHello(cfg.ServPref, text) {
		helloList := toCommandHelloChats(cfg.Hello)
		commands.HandleRemoveHello(vk, msg.PeerID, msg.MessageID, cfg.ID, msg.FromID, helloList, func(newList []commands.HelloChat) {
			cfg.Hello = fromCommandHelloChats(newList)
			cfg.Save()
		})
		return
	}

	// --- Алиас-триггер ---
	alias := matchAlias(cfg, text)
	if alias != nil {
		signal, separator := parseAliasSignal(text, alias.SokrCmd)
		commands.HandleAliasSignal(vk, msg.PeerID, msg.MessageID, msg.ConversationMessageID,
			msg.FromID, cfg.ID, msg.Timestamp, alias.PolnCmd, signal, separator,
			cfg.SecretCode, callbackLink(), cfg.Token)
		return
	}

	// --- Основной сигнал ---
	if cfg.HasMyPrefix(text) {
		pref := cfg.GetMyPrefix(text)
		signalText := strings.TrimSpace(text[len(pref):])
		if signalText != "" {
			commands.HandleSendSignal(vk, msg.PeerID, msg.MessageID, msg.ConversationMessageID,
				msg.FromID, cfg.ID, msg.Timestamp, signalText, cfg.SecretCode, callbackLink())
		}
		return
	}
}

func middleware(vk *api.VK, cfg *UserConfig, msg *MessageEvent) bool {
	if msg.FromID < 0 {
		return false
	}

	if msg.FromID == cfg.ID {
		return true
	}

	if cfg.IsDov(msg.FromID) {
		words := strings.SplitN(msg.Text, " ", 2)
		if len(words) >= 2 && strings.ToLower(words[0]) == strings.ToLower(cfg.DovPref) {
			vk.MessagesSend(api.Params{
				"peer_id":   msg.PeerID,
				"message":   words[1],
				"random_id": 0,
			})
			return false
		}
	}

	if cfg.IsGlobalIgnored(msg.FromID) {
		vk.Request("messages.delete", api.Params{
			"peer_id":        msg.PeerID,
			"message_ids":    msg.MessageID,
			"delete_for_all": 1,
		})
		return false
	}

	if cfg.IsChatIgnored(msg.FromID, msg.PeerID) {
		vk.Request("messages.delete", api.Params{
			"peer_id":        msg.PeerID,
			"message_ids":    msg.MessageID,
			"delete_for_all": 1,
		})
		return false
	}

	return false
}

func handleChatInvite(vk *api.VK, cfg *UserConfig, msg *MessageEvent) {
	helloList := toCommandHelloChats(cfg.Hello)
	commands.HandleChatInvite(vk, msg.PeerID, helloList)
}

func matchAlias(cfg *UserConfig, text string) *Alias {
	lower := strings.ToLower(text)
	for i, a := range cfg.Aliases {
		sokr := strings.ToLower(a.SokrCmd)
		if lower == sokr || strings.HasPrefix(lower, sokr+" ") || strings.HasPrefix(lower, sokr+"\n") {
			return &cfg.Aliases[i]
		}
	}
	return nil
}

func parseAliasSignal(text, sokrCmd string) (signal, separator string) {
	rest := strings.TrimSpace(text[len(sokrCmd):])
	if rest == "" {
		return "", " "
	}
	raw := text[len(sokrCmd):]
	if strings.HasPrefix(raw, "\n") {
		return strings.TrimPrefix(raw, "\n"), "\n"
	}
	return strings.TrimPrefix(raw, " "), " "
}

func toCommandAliases(in []Alias) []commands.Alias {
	out := make([]commands.Alias, len(in))
	for i, a := range in {
		out[i] = commands.Alias{Name: a.Name, SokrCmd: a.SokrCmd, PolnCmd: a.PolnCmd}
	}
	return out
}

func fromCommandAliases(in []commands.Alias) []Alias {
	out := make([]Alias, len(in))
	for i, a := range in {
		out[i] = Alias{Name: a.Name, SokrCmd: a.SokrCmd, PolnCmd: a.PolnCmd}
	}
	return out
}

func toCommandIgnoreUsers(in []IgnoreUser) []commands.IgnoreUser {
	out := make([]commands.IgnoreUser, len(in))
	for i, ig := range in {
		out[i] = commands.IgnoreUser{UserID: ig.UserID, PeerID: ig.PeerID}
	}
	return out
}

func fromCommandIgnoreUsers(in []commands.IgnoreUser) []IgnoreUser {
	out := make([]IgnoreUser, len(in))
	for i, ig := range in {
		out[i] = IgnoreUser{UserID: ig.UserID, PeerID: ig.PeerID}
	}
	return out
}

func toCommandHelloChats(in []HelloChat) []commands.HelloChat {
	out := make([]commands.HelloChat, len(in))
	for i, h := range in {
		out[i] = commands.HelloChat{PeerID: h.PeerID, HelloText: h.HelloText}
	}
	return out
}

func fromCommandHelloChats(in []commands.HelloChat) []HelloChat {
	out := make([]HelloChat, len(in))
	for i, h := range in {
		out[i] = HelloChat{PeerID: h.PeerID, HelloText: h.HelloText}
	}
	return out
}
