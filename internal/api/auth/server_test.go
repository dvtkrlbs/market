package auth

import (
	"net/http/httptest"
	"testing"
)

func TestForwardedIPPrecedence(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "192.0.2.3:1234"
	request.Header.Set("X-Forwarded-For", "198.51.100.2, 192.0.2.3")
	request.Header.Set("Forwarded", `for="[2001:db8::1]:4711";proto=https`)

	got := forwardedIP(request)
	if got == nil || *got != "2001:db8::1" {
		t.Fatalf("forwardedIP() = %v, want 2001:db8::1", got)
	}
}

func TestForwardedIPFallsBackToRemoteAddress(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "192.0.2.3:1234"

	got := forwardedIP(request)
	if got == nil || *got != "192.0.2.3" {
		t.Fatalf("forwardedIP() = %v, want 192.0.2.3", got)
	}
}
