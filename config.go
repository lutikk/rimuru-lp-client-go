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

	"github.com/SevereCloud/vksdk/v3/api"

	"rimuru_lp_client_go/lp"
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

// baseDir — директория рядом с исполняемым файлом. Именно сюда кладём config.json,
// чтобы клиент можно было запускать из любого места (cd в другую папку, ярлык,
// systemd/cron с чужим WorkingDirectory) и он всегда находил свой конфиг.
// Фолбэк на текущую рабочую директорию — только если путь к бинарнику недоступен
// (крайне редкий случай) или мы под `go run` (бинарник во временной папке).
func baseDir() string {
	exe, err := os.Executable()
	if err == nil {
		if resolved, lerr := filepath.EvalSymlinks(exe); lerr == nil {
			exe = resolved
		}
		if dir := filepath.Dir(exe); !isGoRunTemp(dir) {
			return dir
		}
	}
	dir, _ := os.Getwd()
	return dir
}

// isGoRunTemp распознаёт временную директорию `go run` (go-build...),
// чтобы при разработке конфиг не терялся во временной папке, а падал в cwd.
func isGoRunTemp(dir string) bool {
	return strings.Contains(dir, "go-build") ||
		strings.HasPrefix(dir, os.TempDir())
}

func configPath() string {
	return filepath.Join(baseDir(), "config.json")
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
	// Защита от «отравленного» конфига: старые версии на silent-ссылке сохраняли
	// в поле token весь URL целиком, из-за чего клиент падал при каждом запуске,
	// пока config.json не удалят вручную. Теперь такой токен ловим и переспрашиваем.
	if looksBrokenToken(cfg.Token) {
		fmt.Println("⚠ Сохранённый токен выглядит некорректным (не похож на VK access_token).")
		fmt.Println("  Повторная настройка.")
		fmt.Println()
		return initNewConfig()
	}
	return &cfg
}

func initNewConfig() *UserConfig {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Файл конфигурации не найден — нужна первичная настройка.")
	fmt.Println()
	fmt.Println("1) Открой в браузере эту ссылку и войди в свой VK-аккаунт:")
	fmt.Println()
	fmt.Println("   " + lp.AuthURL)
	fmt.Println()
	fmt.Println("2) После входа тебя перекинет на страницу oauth.vk.com/blank.html —")
	fmt.Println("   скопируй ВЕСЬ адрес из строки браузера (в нём есть #access_token=...)")
	fmt.Println("   и вставь сюда. Можно вставить и сам токен (vk1.a....).")
	fmt.Println()

	var token string
	var userID int
	for {
		fmt.Print("Вставь ссылку с токеном или сам токен: ")
		tokenInput, _ := reader.ReadString('\n')
		tokenInput = strings.TrimSpace(tokenInput)

		if tokenInput == "" {
			fmt.Println("  Пустой ввод. Попробуй ещё раз.")
			continue
		}

		// Отсекаем silent_token (VK ID): им НЕЛЬЗЯ пользоваться как access_token,
		// а обменять его без сервисного ключа приложения-инициатора невозможно.
		if isSilentToken(tokenInput) {
			fmt.Println("  ✗ Это silent-токен VK ID (в ссылке #payload=...\"type\":\"silent_token\"),")
			fmt.Println("    а не access_token — использовать напрямую нельзя.")
			fmt.Println("    Открой ссылку выше именно в обычном браузере (не через VK ID one-tap)")
			fmt.Println("    и скопируй адрес страницы blank.html с #access_token=...")
			continue
		}

		token, userID = parseToken(tokenInput)

		// Валидируем токен ДО сохранения (users.get). Заодно добираем user_id,
		// если во вводе была не ссылка, а голый токен.
		id, err := validateToken(token)
		if err != nil {
			fmt.Printf("  ✗ Токен не прошёл проверку VK (users.get): %v\n", err)
			fmt.Println("    Убедись, что скопировал именно access_token, и попробуй снова.")
			continue
		}
		if id != 0 {
			userID = id
		}
		break
	}

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
	fmt.Printf("✓ Токен принят (user_id=%d), конфиг сохранён.\n\n", userID)
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

// isSilentToken распознаёт ссылку/строку VK ID с silent-токеном
// (redirect на blank.html#payload=...{"type":"silent_token"...}).
func isSilentToken(input string) bool {
	if !strings.Contains(input, "#") {
		return false
	}
	fragment := strings.SplitN(input, "#", 2)[1]
	if params, err := url.ParseQuery(fragment); err == nil {
		if payload := params.Get("payload"); payload != "" {
			return strings.Contains(payload, "silent_token")
		}
	}
	// Фолбэк: payload мог не распарситься как query — ищем маркер в сырой строке.
	return strings.Contains(input, "silent_token")
}

// validateToken проверяет VK access_token через users.get и возвращает id владельца.
// Пустой/битый токен или ошибка VK => токен не принимается.
func validateToken(token string) (int, error) {
	if token == "" {
		return 0, fmt.Errorf("пустой токен")
	}
	users, err := api.NewVK(token).UsersGet(api.Params{})
	if err != nil {
		return 0, err
	}
	if len(users) == 0 {
		return 0, fmt.Errorf("users.get вернул пустой ответ")
	}
	return users[0].ID, nil
}

// looksBrokenToken — грубая проверка, что в поле token лежит НЕ VK access_token,
// а мусор (URL/фрагмент). Настоящий токен vk1.a.... не содержит "://" и "#".
func looksBrokenToken(token string) bool {
	t := strings.TrimSpace(token)
	return t == "" || strings.Contains(t, "://") || strings.Contains(t, "#")
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
