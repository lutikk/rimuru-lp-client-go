package main

// Go-хаб (порт rimuruduty на Go). Go-клиент работает ТОЛЬКО с Go-API.
const baseDomain = "https://go.rimuruproject.ru"

func callbackLink() string {
	return baseDomain + "/callback/"
}

func getLPInfoLink() string {
	return baseDomain + "/secret_code/"
}

func versionLink() string {
	return baseDomain + "/lp/version/"
}
