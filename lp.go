package main

// Go-хаб (порт rimuruduty на Go). Go-клиент работает ТОЛЬКО с Go-API.
const baseDomain = "https://go.rimuruproject.ru"

// vkAuthURL — ссылка для получения VK access_token через официальное приложение
// VK Android (app_id 2274003), implicit-флоу. Уходим с Kate Mobile (2685278) из-за
// участившихся ограничений/заморозок на его токенах. Юзер открывает ссылку в браузере,
// логинится, и его редиректит на https://oauth.vk.com/blank.html#access_token=vk1.a....&user_id=...
// scope=messages,offline — минимум для LP-юзербота + непротухающий токен.
const vkAuthURL = "https://oauth.vk.com/authorize?client_id=2274003&scope=messages,offline&redirect_uri=https://oauth.vk.com/blank.html&display=page&response_type=token&revoke=1&v=5.199"

func callbackLink() string {
	return baseDomain + "/callback/"
}

func getLPInfoLink() string {
	return baseDomain + "/secret_code/"
}

func versionLink() string {
	return baseDomain + "/lp/version/"
}
