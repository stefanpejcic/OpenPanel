package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseJWT(t *testing.T) {
	secret := []byte("test-secret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "42",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("signing test token: %v", err)
	}

	userID, err := parseJWT(signed, secret)
	if err != nil {
		t.Fatalf("parseJWT: %v", err)
	}
	if userID != 42 {
		t.Errorf("userID = %d, want 42", userID)
	}
}

func TestParseJWTWrongSecret(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1"})
	signed, err := token.SignedString([]byte("secret-a"))
	if err != nil {
		t.Fatalf("signing test token: %v", err)
	}

	if _, err := parseJWT(signed, []byte("secret-b")); err == nil {
		t.Error("expected an error validating a token against the wrong secret")
	}
}

func TestParseJWTExpired(t *testing.T) {
	secret := []byte("test-secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "1",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	signed, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("signing test token: %v", err)
	}

	if _, err := parseJWT(signed, secret); err == nil {
		t.Error("expected an error for an expired token")
	}
}

func TestParseJWTRejectsAlgNone(t *testing.T) {
	// alg=none is the classic JWT confusion attack: an attacker-crafted
	// token with no signature that a naive verifier accepts anyway.
	unsigned := "eyJhbGciOiJub25lIn0.eyJzdWIiOiIxIn0."
	if _, err := parseJWT(unsigned, []byte("test-secret")); err == nil {
		t.Error("expected alg=none token to be rejected")
	}
}
