package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

// CookieSigner authenticates session cookie values without encrypting them.
type CookieSigner struct {
	key []byte
}

func NewCookieSigner(key []byte) (*CookieSigner, error) {
	if len(key) < 32 {
		return nil, fmt.Errorf("session cookie signing key must be at least 32 bytes")
	}
	return &CookieSigner{key: append([]byte(nil), key...)}, nil
}

func (s *CookieSigner) Sign(token string) string {
	return token + "." + base64.RawURLEncoding.EncodeToString(s.signature(token))
}

func (s *CookieSigner) Verify(value string) (string, error) {
	index := strings.LastIndexByte(value, '.')
	if index <= 0 {
		return "", ErrInvalidSession
	}
	token := value[:index]
	signature, err := base64.RawURLEncoding.Strict().DecodeString(value[index+1:])
	if err != nil || !hmac.Equal(signature, s.signature(token)) {
		return "", ErrInvalidSession
	}
	return token, nil
}

func (s *CookieSigner) signature(token string) []byte {
	mac := hmac.New(sha256.New, s.key)
	// Bind the signature to this cookie's purpose.
	mac.Write([]byte("market:session:"))
	mac.Write([]byte(token))
	return mac.Sum(nil)
}
