package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Alias struct {
	Name    string `json:"name"`
	SokrCmd string `json:"sokr_cmd"`
	PolnCmd string `json:"poln_cmd"`
}

type IgnoreUser struct {
	UserID int `json:"user_id"`
	PeerID int `json:"peer_id"`
}

type HelloChat struct {
	PeerID    int    `json:"peer_id"`
	HelloText string `json:"hello_text"`
}

type UserConfig struct {
	mu           sync.RWMutex
	ID           int          `json:"id"`
	Token        string       `json:"token"`
	MyPref       []string     `json:"my_pref"`
	ServPref     []string     `json:"serv_pref"`
	TimeCommands float64      `json:"time_commands"`
	Aliases      []Alias      `json:"alias"`
	Dov          []int        `json:"dov"`
	DovPref      string       `json:"dov_pref"`
	SecretCode   string       `json:"secret_code"`
	GlobalIgnore []int        `json:"global_ignore"`
	Ignore       []IgnoreUser `json:"ignore"`
	Hello        []HelloChat  `json:"hello"`
}

func configPath() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "config.json")
}

func (u *UserConfig) Save() error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	data, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return err
	}
	// 0600: файл содержит VK-токен пользователя — доступ только владельцу.
	return os.WriteFile(configPath(), data, 0600)
}

func LoadConfig() *UserConfig {
	data, err := os.ReadFile(configPath())
	if err != nil || len(data) == 0 {
		return initNewConfig()
	}

	var cfg UserConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return initNewConfig()
	}
	return &cfg
}

func initNewConfig() *UserConfig {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Файл конфигурации не найден. Пожалуйста, введите токен: ")
	tokenInput, _ := reader.ReadString('\n')
	tokenInput = strings.TrimSpace(tokenInput)

	token, userID := parseToken(tokenInput)

	cfg := &UserConfig{
		ID:           userID,
		Token:        token,
		MyPref:       []string{".л", "!л"},
		ServPref:     []string{".слп ", "!слп "},
		TimeCommands: 8000,
		DovPref:      "..",
		SecretCode:   "Noy",
	}
	cfg.Save()
	return cfg
}

func parseToken(input string) (string, int) {
	if strings.Contains(input, "#") {
		fragment := strings.SplitN(input, "#", 2)[1]
		params, err := url.ParseQuery(fragment)
		if err == nil {
			token := params.Get("access_token")
			uid, _ := strconv.Atoi(params.Get("user_id"))
			if token != "" && uid != 0 {
				return token, uid
			}
		}
	}
	return input, 0
}

func (u *UserConfig) HasMyPrefix(text string) bool {
	lower := strings.ToLower(text)
	for _, p := range u.MyPref {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

func (u *UserConfig) GetMyPrefix(text string) string {
	lower := strings.ToLower(text)
	for _, p := range u.MyPref {
		if strings.HasPrefix(lower, p) {
			return p
		}
	}
	return ""
}

func (u *UserConfig) HasServPrefix(text string) bool {
	lower := strings.ToLower(text)
	for _, p := range u.ServPref {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

func (u *UserConfig) GetServPrefix(text string) string {
	lower := strings.ToLower(text)
	for _, p := range u.ServPref {
		if strings.HasPrefix(lower, p) {
			return p
		}
	}
	return ""
}

func (u *UserConfig) MatchAlias(text string) *Alias {
	lower := strings.ToLower(text)
	for i, a := range u.Aliases {
		if strings.HasPrefix(lower, strings.ToLower(a.SokrCmd)) {
			return &u.Aliases[i]
		}
	}
	return nil
}

func (u *UserConfig) IsDov(userID int) bool {
	for _, id := range u.Dov {
		if id == userID {
			return true
		}
	}
	return false
}

func (u *UserConfig) IsGlobalIgnored(userID int) bool {
	for _, id := range u.GlobalIgnore {
		if id == userID {
			return true
		}
	}
	return false
}

func (u *UserConfig) IsChatIgnored(userID, peerID int) bool {
	for _, ig := range u.Ignore {
		if ig.UserID == userID && ig.PeerID == peerID {
			return true
		}
	}
	return false
}
