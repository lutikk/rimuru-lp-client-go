// Package lp — общие константы и версия LP-клиента. Вынесены в отдельный пакет,
// чтобы их могли импортировать И main, И commands (пакет main импортировать нельзя).
package lp

// Version — версия сборки. Дефолт перезаписывается линкером на релизе:
//
//	-ldflags "-X rimuru_lp_client_go/lp.Version=vX.Y.Z"
var Version = "5.0.2"

// BaseDomain — Go-хаб (порт rimuruduty на Go). Go-клиент работает ТОЛЬКО с Go-API.
const BaseDomain = "https://go.rimuruproject.ru"

// AuthURL — ссылка получения VK access_token, classic implicit-флоу.
// ТЕКУЩИЙ КАНДИДАТ: Маруся (app_id 6463690) — сайт раньше через неё слал messages.send
// (работало) → она реально ДАЁТ messages, и душат её меньше Kate. Classic-форма
// oauth.vk.com/authorize (НЕ id.vk.com!) отдаёт прямой #access_token, а не silent.
// Euphoria 4510232 отвалилась — её токен без messages (502 на приветствии). Фолбэк — Kate 2685278.
// Проверка scope выданного токена: account.getAppPermissions (messages = бит 4096).
const AuthURL = "https://oauth.vk.com/authorize?client_id=6463690&scope=messages,offline&redirect_uri=https://oauth.vk.com/blank.html&display=page&response_type=token&revoke=1&v=5.199"

// CallbackLink — приём сигналов дежурного (lpSendMySignal).
func CallbackLink() string {
	return BaseDomain + "/callback/"
}

// GetLPInfoLink — выдача secret_code по токену.
func GetLPInfoLink() string {
	return BaseDomain + "/secret_code/"
}

// VersionLink — актуальная версия LP на хабе (для проверки обновлений в .инфо).
func VersionLink() string {
	return BaseDomain + "/lp/version/"
}
