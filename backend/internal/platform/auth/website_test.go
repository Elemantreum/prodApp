package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"testing"
	"time"
)

func TestWebsiteLoginSignatureAndAge(t *testing.T) {
	now := time.Unix(1700000000, 0)
	key := sha256.Sum256([]byte("test-token"))
	mac := hmac.New(sha256.New, key[:])
	mac.Write([]byte("auth_date=1700000000\nfirst_name=Test\nid=42"))
	fields := url.Values{"id": {"42"}, "first_name": {"Test"}, "auth_date": {"1700000000"}, "hash": {hex.EncodeToString(mac.Sum(nil))}}
	auth := NewTelegramAuthenticator("test-token")
	p, err := auth.websitePrincipal(fields, now)
	if err != nil || p.ExternalId != "42" || p.Provider != "telegram" {
		t.Fatalf("unexpected principal: %v, %v", p, err)
	}
	if _, err := auth.websitePrincipal(fields, now.Add(6*time.Minute)); err == nil {
		t.Fatal("accepted expired login")
	}
	fields.Set("id", "43")
	if _, err := auth.websitePrincipal(fields, now); err == nil {
		t.Fatal("accepted modified identity")
	}
}
