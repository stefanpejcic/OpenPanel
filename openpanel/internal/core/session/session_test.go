package session

import (
	"net/http"
	"testing"
	"time"
)

func TestNewStoreOptions(t *testing.T) {
	store := NewStore([]byte("test-secret-key-0123456789abcdef"))

	if store.Options.Path != "/" {
		t.Errorf("Path = %q, want /", store.Options.Path)
	}
	if !store.Options.HttpOnly {
		t.Error("HttpOnly = false, want true")
	}
	if store.Options.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", store.Options.SameSite)
	}
	if want := int(300 * time.Minute / time.Second); store.Options.MaxAge != want {
		t.Errorf("MaxAge = %d, want %d", store.Options.MaxAge, want)
	}
	if CookieName != "OPENPANEL" {
		t.Errorf("CookieName = %q, want OPENPANEL", CookieName)
	}
}
