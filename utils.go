package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"rimuru_lp_client_go/lp"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func getCode(token string) (string, error) {
	body, _ := json.Marshal(map[string]string{"token": token})
	// GET с JSON-телом — как в питон-клиенте (requests.get(url, json={'token':...})).
	// Хаб отдаёт /secret_code/ ИМЕННО как GET; POST даёт 405. Раньше клиент слал POST —
	// ошибка порта, из-за неё secret_code не приходил и все callback-сигналы отвергались.
	req, err := http.NewRequest(http.MethodGet, lp.GetLPInfoLink(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("не удалось собрать запрос secret_code: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("не удалось получить secret_code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("хаб вернул статус %d на /secret_code/ (ожидался 200)", resp.StatusCode)
	}

	var result struct {
		SecretCode string `json:"secret_code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа: %w", err)
	}
	if result.SecretCode == "" {
		return "", fmt.Errorf("хаб вернул пустой secret_code")
	}
	return result.SecretCode, nil
}
