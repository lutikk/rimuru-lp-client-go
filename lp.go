package main

const baseDomain = "https://api.rimuruproject.ru"

func callbackLink() string {
	return baseDomain + "/callback/"
}

func getLPInfoLink() string {
	return baseDomain + "/secret_code/"
}

func versionLink() string {
	return baseDomain + "/lp/version/"
}
