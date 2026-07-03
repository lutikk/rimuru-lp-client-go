package main

import "testing"

// Реальный silent-payload из репорта пользователя (укорочен, структура та же).
const realSilentURL = `https://oauth.vk.com/blank.html#payload=%7B%22type%22%3A%22silent_token%22%2C%22auth%22%3A1%2C%22user%22%3A%7B%22id%22%3A812747657%7D%2C%22token%22%3A%22PAdmD_caqHo9%22%2C%22ttl%22%3A600%2C%22uuid%22%3A%22%22%2C%22hash%22%3A%22y8lViY9%22%7D`

const normalTokenURL = `https://oauth.vk.com/blank.html#access_token=vk1.a.AbCdEf&expires_in=0&user_id=812747657`

func TestIsSilentToken(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"real silent payload", realSilentURL, true},
		{"normal access_token link", normalTokenURL, false},
		{"raw token", "vk1.a.AbCdEf", false},
		{"empty", "", false},
	}
	for _, c := range cases {
		if got := isSilentToken(c.in); got != c.want {
			t.Errorf("isSilentToken(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestParseToken(t *testing.T) {
	tok, uid := parseToken(normalTokenURL)
	if tok != "vk1.a.AbCdEf" || uid != 812747657 {
		t.Errorf("parseToken(normal) = (%q, %d), want (vk1.a.AbCdEf, 812747657)", tok, uid)
	}
	// silent-ссылка: access_token нет => отдаёт всю строку и uid=0 (её отсечёт isSilentToken до parseToken).
	tok2, uid2 := parseToken(realSilentURL)
	if uid2 != 0 {
		t.Errorf("parseToken(silent) uid = %d, want 0", uid2)
	}
	_ = tok2
	// голый токен => как есть, uid=0 (добьём через users.get).
	tok3, uid3 := parseToken("vk1.a.Raw")
	if tok3 != "vk1.a.Raw" || uid3 != 0 {
		t.Errorf("parseToken(raw) = (%q, %d), want (vk1.a.Raw, 0)", tok3, uid3)
	}
}

func TestLooksBrokenToken(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{realSilentURL, true},
		{"https://oauth.vk.com/blank.html", true},
		{"", true},
		{"vk1.a.AbCdEf", false},
		{"a1b2c3d4e5f6", false}, // старый hex-формат
	}
	for _, c := range cases {
		if got := looksBrokenToken(c.in); got != c.want {
			t.Errorf("looksBrokenToken(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
