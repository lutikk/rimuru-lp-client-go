package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func getCode(token string) (string, error) {
	body, _ := json.Marshal(map[string]string{"token": token})
	resp, err := httpClient.Post(getLPInfoLink(), "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("не удалось получить secret_code: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		SecretCode string `json:"secret_code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа: %w", err)
	}
	return result.SecretCode, nil
}
