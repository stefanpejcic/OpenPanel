package account

import (
	"encoding/base32"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestRandomBase32Secret(t *testing.T) {
	secret, err := randomBase32Secret()
	if err != nil {
		t.Fatalf("randomBase32Secret: %v", err)
	}
	if len(secret) != 32 {
		t.Errorf("expected a 32-character secret, got %d: %q", len(secret), secret)
	}
	if _, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret); err != nil {
		t.Errorf("expected valid base32, got decode error: %v", err)
	}

	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateCode with generated secret: %v", err)
	}
	if !totp.Validate(code, secret) {
		t.Error("expected a code generated from the fresh secret to validate against it")
	}
}

func TestNullableString(t *testing.T) {
	if nullableString("") != nil {
		t.Error("expected nil for an empty string")
	}
	if nullableString("abc") != "abc" {
		t.Error("expected the string itself when non-empty")
	}
}
