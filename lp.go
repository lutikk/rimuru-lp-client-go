package main

// Go-хаб (порт rimuruduty на Go). Go-клиент работает ТОЛЬКО с Go-API.
const baseDomain = "https://go.rimuruproject.ru"

// vkAuthURL — ссылка получения VK access_token через Kate Mobile (app_id 2685278),
// classic implicit-флоу. VK Android 2274003 НЕ подошёл: он direct-auth (вход логин/пароль),
// и VK запрещает ему браузерный authorize?response_type=token ("Unavailable for apps with
// direct auth"). Kate — единственное надёжное implicit-приложение в 2026: отдаёт прямой
// #access_token=vk1.a....&user_id=... на blank.html с messages-scope.
// Минус: за Kate VK душит аккаунты за автоматизацию — durable-выход = своё VK ID app + PKCE.
const vkAuthURL = "https://oauth.vk.com/authorize?client_id=2685278&scope=messages,offline&redirect_uri=https://oauth.vk.com/blank.html&display=page&response_type=token&revoke=1&v=5.199"

func callbackLink() string {
	return baseDomain + "/callback/"
}

func getLPInfoLink() string {
	return baseDomain + "/secret_code/"
}

func versionLink() string {
	return baseDomain + "/lp/version/"
}
