package main

// Go-хаб (порт rimuruduty на Go). Go-клиент работает ТОЛЬКО с Go-API.
const baseDomain = "https://go.rimuruproject.ru"

// vkAuthURL — ссылка получения VK access_token, classic implicit-флоу.
// ТЕКУЩИЙ КАНДИДАТ: Euphoria (app_id 4510232) — сторонний нишевый implicit-клиент,
// проверяем гипотезу «душат меньше Kate». Фолбэк — Kate Mobile 2685278 (надёжный, но
// душат сильнее). Если Euphoria не даёт messages/ловит silent — перебираем следующего
// кандидата (Phoenix 4994316, Zeus 4831060, Rocket 4757672, VK MD 4967124, Маруся 6463690).
// Проверка scope выданного токена: account.getAppPermissions.
const vkAuthURL = "https://oauth.vk.com/authorize?client_id=4510232&scope=messages,offline&redirect_uri=https://oauth.vk.com/blank.html&display=page&response_type=token&revoke=1&v=5.199"

func callbackLink() string {
	return baseDomain + "/callback/"
}

func getLPInfoLink() string {
	return baseDomain + "/secret_code/"
}

func versionLink() string {
	return baseDomain + "/lp/version/"
}
