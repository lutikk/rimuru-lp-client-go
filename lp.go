package main

// Go-хаб (порт rimuruduty на Go). Go-клиент работает ТОЛЬКО с Go-API.
const baseDomain = "https://go.rimuruproject.ru"

// vkAuthURL — ссылка получения VK access_token, classic implicit-флоу.
// ТЕКУЩИЙ КАНДИДАТ: Маруся (app_id 6463690) — сайт раньше через неё слал messages.send
// (работало) → она реально ДАЁТ messages, и душат её меньше Kate. Classic-форма
// oauth.vk.com/authorize (НЕ id.vk.com!) отдаёт прямой #access_token, а не silent.
// Euphoria 4510232 отвалилась — её токен без messages (502 на приветствии). Фолбэк — Kate 2685278.
// Проверка scope выданного токена: account.getAppPermissions (messages = бит 4096).
const vkAuthURL = "https://oauth.vk.com/authorize?client_id=6463690&scope=messages,offline&redirect_uri=https://oauth.vk.com/blank.html&display=page&response_type=token&revoke=1&v=5.199"

func callbackLink() string {
	return baseDomain + "/callback/"
}

func getLPInfoLink() string {
	return baseDomain + "/secret_code/"
}

func versionLink() string {
	return baseDomain + "/lp/version/"
}
