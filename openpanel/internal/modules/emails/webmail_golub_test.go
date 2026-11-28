package emails

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGolubSSOToken(t *testing.T) {
	secret := strings.Repeat("k", 64)
	now := time.Unix(1_800_000_000, 0)
	token, err := golubSSOToken(secret, "ana@example.com", now)
	if err != nil {
		t.Fatal(err)
	}
	p64, m64, ok := strings.Cut(token, ".")
	if !ok {
		t.Fatalf("token %q", token)
	}
	payload, _ := base64.RawURLEncoding.DecodeString(p64)
	sig, _ := base64.RawURLEncoding.DecodeString(m64)
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(payload)
	if !hmac.Equal(sig, m.Sum(nil)) {
		t.Fatal("signature does not verify")
	}
	parts := strings.Split(string(payload), "|")
	if len(parts) != 5 || parts[0] != "ana@example.com" || len(parts[2]) < 16 {
		t.Fatalf("payload %q", payload)
	}
	if exp, _ := strconv.ParseInt(parts[1], 10, 64); exp != now.Add(golubTokenTTL).Unix() || golubTokenTTL > time.Minute {
		t.Errorf("expires %s", parts[1])
	}
	if again, _ := golubSSOToken(secret, "ana@example.com", now); again == token {
		t.Error("each token needs its own nonce")
	}
	if _, err := golubSSOToken("short", "ana@example.com", now); err == nil {
		t.Error("a short key must be refused")
	}
}

func TestWebmailFromNames(t *testing.T) {
	for names, want := range map[string]string{
		"openpanel\nopenadmin_golub\n":     "golub",
		"openadmin_roundcube\nopenpanel\n": "roundcube",
		"golub-bench-roundcube\nopenpanel": "",
		"":                                 "",
	} {
		if got := webmailFromNames(names); got != want {
			t.Errorf("%q: %q", names, got)
		}
	}
}
