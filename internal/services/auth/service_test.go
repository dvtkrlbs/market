package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestSessionTokenPartsAndDigest(t *testing.T) {
	id, err := randomTokenPart(16)
	if err != nil {
		t.Fatal(err)
	}
	secret, err := randomTokenPart(32)
	if err != nil {
		t.Fatal(err)
	}
	gotID, gotSecret, err := parseToken(id + "." + secret)
	if err != nil {
		t.Fatal(err)
	}
	if gotID != id || gotSecret != secret {
		t.Fatalf("parsed token = %q.%q, want original", gotID, gotSecret)
	}
	if digest := secretDigest(secret); digest == secret || len(digest) != 43 {
		t.Fatalf("secret digest = %q, want 32-byte base64url digest", digest)
	}
}

func TestSessionTokenRejectsMalformedValues(t *testing.T) {
	for _, token := range []string{"", "missing-dot", ".secret", "id.", "a.b.c", "short.c2hvcnQ"} {
		if _, _, err := parseToken(token); !errors.Is(err, ErrInvalidSession) {
			t.Errorf("parseToken(%q) error = %v, want invalid session", token, err)
		}
	}
}

func TestNormalizeEmail(t *testing.T) {
	got, err := normalizeEmail("  Person@Example.COM ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "person@example.com" {
		t.Fatalf("email = %q", got)
	}
	for _, value := range []string{"", "not-an-email", strings.Repeat("a", 321)} {
		if _, err := normalizeEmail(value); !errors.Is(err, ErrInvalidEmail) {
			t.Errorf("normalizeEmail(%q) error = %v, want invalid email", value, err)
		}
	}
}
