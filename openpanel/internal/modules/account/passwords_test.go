package account

import "testing"

func TestIsCyberPanelHash(t *testing.T) {
	sixtyFourChars := make([]byte, 64)
	for i := range sixtyFourChars {
		sixtyFourChars[i] = 'a'
	}
	valid := string(sixtyFourChars) + ":somesalt"

	if !isCyberPanelHash(valid) {
		t.Errorf("expected %q to look like a CyberPanel hash", valid)
	}
	if isCyberPanelHash("short:salt") {
		t.Error("expected a short hash part to be rejected")
	}
	if isCyberPanelHash("no-colon-here") {
		t.Error("expected a hash with no colon to be rejected")
	}
}

func TestVerifyCyberPanelPassword(t *testing.T) {
	salt := "0123456789abcdef0123456789abcdef" // 32 chars
	salt = salt[:32]
	hash := sha256Hex("mypassword" + salt)
	stored := hash + ":" + salt

	if !verifyCyberPanelPassword("mypassword", stored) {
		t.Error("expected correct password to verify")
	}
	if verifyCyberPanelPassword("wrongpassword", stored) {
		t.Error("expected incorrect password to fail")
	}
}

func TestVerifyCyberPanelPasswordMalformed(t *testing.T) {
	if verifyCyberPanelPassword("x", "no-colon") {
		t.Error("expected malformed hash (no colon) to fail closed")
	}
	if verifyCyberPanelPassword("x", "short:short") {
		t.Error("expected malformed hash (wrong lengths) to fail closed")
	}
}
