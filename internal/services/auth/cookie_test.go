package auth

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestCookieSignerRoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{1}, 32)
	signer, err := NewCookieSigner(key)
	if err != nil {
		t.Fatal(err)
	}
	// The signer must not retain caller-owned mutable key storage.
	key[0] = 2
	other, err := NewCookieSigner(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	token := "session-id.session-secret"
	signed := signer.Sign(token)
	if signed == token {
		t.Fatal("cookie is unsigned")
	}
	got, err := other.Verify(signed)
	if err != nil || got != token {
		t.Fatalf("Verify() = %q, %v, want original token", got, err)
	}
}

func TestCookieSignerRejectsInvalidValues(t *testing.T) {
	signer, err := NewCookieSigner(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	other, err := NewCookieSigner(bytes.Repeat([]byte{2}, 32))
	if err != nil {
		t.Fatal(err)
	}
	signed := signer.Sign("session-id.session-secret")
	for _, value := range []string{
		"", "unsigned", "session-id.session-secret", signed + "extra",
		strings.Replace(signed, "session-id", "modified-id", 1),
		signed[:len(signed)-1], other.Sign("session-id.session-secret"),
		"session-id.session-secret.%%%", ".signature",
	} {
		if _, err := signer.Verify(value); !errors.Is(err, ErrInvalidSession) {
			t.Errorf("Verify(%q) error = %v, want invalid session", value, err)
		}
	}
}

func TestCookieSignerRequiresStrongKey(t *testing.T) {
	for _, length := range []int{0, 16, 31} {
		if _, err := NewCookieSigner(make([]byte, length)); err == nil {
			t.Errorf("accepted %d-byte key", length)
		}
	}
}
