package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestPasswordPHCRoundTrip(t *testing.T) {
	encoded, err := hashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=19$") {
		t.Fatalf("hash = %q, want Argon2id PHC string", encoded)
	}
	valid, err := verifyPassword(encoded, "correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Fatal("correct password did not verify")
	}
	valid, err = verifyPassword(encoded, "wrong password")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Fatal("wrong password verified")
	}
}

func TestPasswordRejectsMalformedOrExcessivePHCParameters(t *testing.T) {
	tests := []string{
		"not-a-phc-string",
		"$argon2id$v=18$m=65536,t=3,p=2$c2FsdHNhbHRzYWx0c2FsdA$ZmFrZWZha2VmYWtlZmFrZWZha2VmYWtlZmFrZQ",
		"$argon2id$v=19$m=999999,t=3,p=2$c2FsdHNhbHRzYWx0c2FsdA$ZmFrZWZha2VmYWtlZmFrZWZha2VmYWtlZmFrZQ",
	}
	for _, encoded := range tests {
		if _, err := verifyPassword(encoded, "password"); !errors.Is(err, errInvalidPasswordHash) {
			t.Errorf("verifyPassword(%q) error = %v, want invalid hash", encoded, err)
		}
	}
}
