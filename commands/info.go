package commands

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/shirou/gopsutil/v4/process"
)

const version = "4.6.4"

func MatchInfo(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"инфо") {
			return true
		}
	}
	return false
}

func MatchServer(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"сервер") {
			return true
		}
	}
	return false
}

func MatchAddPref(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"+преф ") {
			return true
		}
	}
	return false
}

func MatchRemovePref(servPref []string, text string) bool {
	lower := strings.ToLower(text)
	for _, p := range servPref {
		if strings.HasPrefix(lower, p+"-преф ") {
			return true
		}
	}
	return false
}

type InfoData struct {
	MyPref       []string
	ServPref     []string
	AliasCount   int
	DovCount     int
	DovPref      string
	IgnoreCount  int
	GIgnoreCount int
	HelloCount   int
}

func HandleInfo(vk *api.VK, peerID, messageID, ownerID, fromID int, data InfoData, versionURL string) {
	if fromID != ownerID {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memMB := float64(m.Alloc) / 1024 / 1024

	text := fmt.Sprintf("RimuruLP v%s (Go)\n", version)
	text += fmt.Sprintf("Мои префиксы: %s\n", strings.Join(data.MyPref, " "))
	text += fmt.Sprintf("Сервисные префиксы: %s\n", strings.Join(data.ServPref, ""))
	text += fmt.Sprintf("Алиасы: %d\n", data.AliasCount)
	text += fmt.Sprintf("Доверенных: %d\n", data.DovCount)
	text += fmt.Sprintf("Повторялка: %s\n", data.DovPref)
	text += fmt.Sprintf("Игнорируем: %d\n", data.IgnoreCount)
	text += fmt.Sprintf("Глобально игнорируем: %d\n", data.GIgnoreCount)
	text += fmt.Sprintf("Приветствия: %d\n", data.HelloCount)
	text += fmt.Sprintf("LP Memory: %.2f Мб\n\n", memMB)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Get(versionURL)
	if err == nil {
		defer resp.Body.Close()
		var vResp struct {
			Version string `json:"version"`
		}
		json.NewDecoder(resp.Body).Decode(&vResp)
		if vResp.Version != "" && vResp.Version != version {
			text += fmt.Sprintf("Вышла новая версия LP %s", vResp.Version)
		}
	}

	EditMessage(vk, peerID, messageID, text)
}

func HandleServer(vk *api.VK, peerID, messageID, ownerID, fromID int) {
	if fromID != ownerID {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memMB := float64(m.Alloc) / 1024 / 1024

	pid := int32(os.Getpid())
	cpuPercent := 0.0
	if proc, err := process.NewProcess(pid); err == nil {
		if pct, err := proc.CPUPercent(); err == nil {
			cpuPercent = pct / float64(runtime.NumCPU())
		}
	}

	text := fmt.Sprintf("LP Memory: %.2f Мб\nLP CPU: %.0f", memMB, cpuPercent)
	EditMessage(vk, peerID, messageID, text)
}

func HandleAddPref(vk *api.VK, peerID, messageID, ownerID, fromID int, text string, servPref, myPref []string, saveFn func([]string)) {
	if fromID != ownerID {
		return
	}

	prefix := getServPrefix(servPref, text)
	sq := strings.TrimSpace(strings.ToLower(text[len(prefix)+len("+преф "):]))

	for _, p := range myPref {
		if p == sq {
			EditMessage(vk, peerID, messageID, fmt.Sprintf("У вас уже есть префикс <<%s>>", sq))
			return
		}
	}

	saveFn(append(myPref, sq))
	EditMessage(vk, peerID, messageID, fmt.Sprintf("Префикс <<%s>> добавлен", sq))
}

func HandleRemovePref(vk *api.VK, peerID, messageID, ownerID, fromID int, text string, servPref, myPref []string, saveFn func([]string)) {
	if fromID != ownerID {
		return
	}

	prefix := getServPrefix(servPref, text)
	sq := strings.TrimSpace(strings.ToLower(text[len(prefix)+len("-преф "):]))

	for i, p := range myPref {
		if p == sq {
			newPref := append(myPref[:i], myPref[i+1:]...)
			saveFn(newPref)
			EditMessage(vk, peerID, messageID, fmt.Sprintf("Префикс <<%s>> удален", sq))
			return
		}
	}

	EditMessage(vk, peerID, messageID, fmt.Sprintf("У вас нет префикса <<%s>>", sq))
}
