package account

import "testing"

func TestBuildNotificationCards(t *testing.T) {
	prefs := []NotificationPref{{Key: "notify_account_login", Value: "1"}, {Key: "notify_account_login_for_known_netblock", Value: "0"}}
	cards := buildNotificationCards(prefs)
	if len(cards) != 12 {
		t.Fatalf("got %d cards, want 12", len(cards))
	}
	if !cards[0].On || len(cards[0].Subs) != 2 || cards[0].Subs[0].On {
		t.Errorf("login card wrong: %+v", cards[0])
	}
	if cards[0].DocsURL != "https://openpanel.com/docs/panel/account/notifications/#new-login" {
		t.Errorf("got %q", cards[0].DocsURL)
	}
}

func TestNotificationDefault(t *testing.T) {
	cases := map[string]int{"notify_account_login": 1, "notify_account_login_for_known_netblock": 0, "notify_username_change": 1, "notify_autossl_expiry": 0}
	for key, want := range cases {
		if got := notificationDefault(key); got != want {
			t.Errorf("%s: got %d want %d", key, got, want)
		}
	}
}

func TestBrowserFromUserAgent(t *testing.T) {
	cases := map[string]string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36":             "Chrome on Windows",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/128.0.0": "Edge on Windows",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15":          "Safari on macOS",
		"Mozilla/5.0 (X11; Linux x86_64; rv:130.0) Gecko/20100101 Firefox/130.0":                                                      "Firefox on Linux",
		"curl/8.5.0": "curl",
		"":           "Unknown",
	}
	for ua, want := range cases {
		if got := browserFromUserAgent(ua); got != want {
			t.Errorf("%q: got %q want %q", ua, got, want)
		}
	}
}

func TestTokenScope(t *testing.T) {
	if got := tokenScope(true, 30); got != " with read-only access that expires in 30 days" {
		t.Errorf("got %q", got)
	}
	if got := tokenScope(false, 0); got != " with full access that never expires" {
		t.Errorf("got %q", got)
	}
}
