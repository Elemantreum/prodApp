package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Principal struct {
	Provider   string
	ExternalId string
	FirstName  string
	LastName   string
	Username   string
}

type Authenticator interface {
	Authenticate(r *http.Request) (*Principal, error)
}

type TelegramAuthenticator struct {
	botToken string
	maxAge   time.Duration
}

func NewTelegramAuthenticator(botToken string) *TelegramAuthenticator {
	return &TelegramAuthenticator{
		botToken: botToken,
		maxAge:   24 * time.Hour,
	}
}

func (a *TelegramAuthenticator) Verify(initData string) (*Principal, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("parse init data: %w", err)
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return nil, errors.New("telegram: missing hash")
	}
	values.Del("hash")

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+values.Get(k))
	}
	dataCheckString := strings.Join(pairs, "\n")

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(a.botToken))
	secretKey := secretMAC.Sum(nil)

	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(dataCheckString))
	computedHash := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(computedHash), []byte(receivedHash)) {
		return nil, errors.New("telegram: invalid hash, possibly forged initData")
	}

	authDateStr := values.Get("auth_date")
	authDateUnix, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("telegram: invalid auth_date: %w", err)
	}
	authDate := time.Unix(authDateUnix, 0)
	if time.Since(authDate) > a.maxAge {
		return nil, errors.New("telegram: initData expired")
	}

	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, errors.New("telegram: missing user data")
	}

	var tgUser struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
	}
	if err := json.Unmarshal([]byte(userJSON), &tgUser); err != nil {
		return nil, fmt.Errorf("telegram: parse user json: %w", err)
	}

	return &Principal{
		Provider:   "telegram",
		ExternalId: strconv.FormatInt(tgUser.ID, 10),
		FirstName:  tgUser.FirstName,
		LastName:   tgUser.LastName,
		Username:   tgUser.Username,
	}, nil
}

type VKAuthenticator struct {
	secretKey string
}

func NewVKAuthenticator(secretKey string) *VKAuthenticator {
	return &VKAuthenticator{secretKey: secretKey}
}

func (a *VKAuthenticator) Verify(rawQuery string) (*Principal, error) {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return nil, fmt.Errorf("vk: parse query: %w", err)
	}

	receivedSign := values.Get("sign")
	if receivedSign == "" {
		return nil, errors.New("vk: missing sign")
	}

	keys := make([]string, 0)
	for k := range values {
		if strings.HasPrefix(k, "vk_") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+values.Get(k))
	}
	dataCheckString := strings.Join(pairs, "&")

	h := hmac.New(sha256.New, []byte(a.secretKey))
	h.Write([]byte(dataCheckString))

	computedSign := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(computedSign), []byte(receivedSign)) {
		return nil, errors.New("vk: invalid sign, possibly forged launch params")
	}

	userID := values.Get("vk_user_id")
	if userID == "" {
		return nil, errors.New("vk: missing vk_user_id")
	}

	return &Principal{
		Provider:   "vk",
		ExternalId: userID,
	}, nil
}
