package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (a *TelegramAuthenticator) websitePrincipal(fields url.Values, now time.Time) (*Principal, error) {
	invalid := errors.New("invalid Telegram login")
	signature, err := hex.DecodeString(fields.Get("hash"))
	if err != nil || len(signature) != sha256.Size {
		return nil, invalid
	}
	pairs := make([]string, 0, len(fields))
	for name, values := range fields {
		if len(values) != 1 || strings.ContainsAny(name, "=\r\n") || strings.ContainsAny(values[0], "\r\n") {
			return nil, invalid
		}
		if name != "hash" {
			pairs = append(pairs, name+"="+values[0])
		}
	}
	sort.Strings(pairs)
	key := sha256.Sum256([]byte(a.botToken))
	mac := hmac.New(sha256.New, key[:])
	mac.Write([]byte(strings.Join(pairs, "\n")))
	if !hmac.Equal(mac.Sum(nil), signature) {
		return nil, invalid
	}
	id, err := strconv.ParseInt(fields.Get("id"), 10, 64)
	if err != nil || id <= 0 {
		return nil, invalid
	}
	stamp, err := strconv.ParseInt(fields.Get("auth_date"), 10, 64)
	age := now.Sub(time.Unix(stamp, 0))
	if err != nil || age < -time.Minute || age > 5*time.Minute {
		return nil, invalid
	}
	return &Principal{Provider: "telegram", ExternalId: strconv.FormatInt(id, 10), FirstName: fields.Get("first_name"), LastName: fields.Get("last_name"), Username: fields.Get("username")}, nil
}
