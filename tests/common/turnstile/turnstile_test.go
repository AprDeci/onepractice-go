package turnstile_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"onepractice-golang/internal/common/turnstile"
)

func TestVerifyDisabledSkips(t *testing.T) {
	v := turnstile.NewVerifier(false, "")

	if err := v.Verify(context.Background(), "", ""); err != nil {
		t.Fatalf("Verify() error = %v, want nil when disabled", err)
	}
}

func TestVerifyNilVerifierSkips(t *testing.T) {
	var v *turnstile.Verifier

	if err := v.Verify(context.Background(), "token", "203.0.113.7"); err != nil {
		t.Fatalf("Verify() error = %v, want nil for nil verifier", err)
	}
}

func TestVerifyEnabledMissingSecret(t *testing.T) {
	v := turnstile.NewVerifier(true, "")

	if err := v.Verify(context.Background(), "token", ""); err == nil {
		t.Fatal("Verify() error = nil, want error when secret key is missing")
	}
}

func TestVerifyEnabledMissingToken(t *testing.T) {
	v := turnstile.NewVerifier(true, "secret")

	if err := v.Verify(context.Background(), "", ""); err == nil {
		t.Fatal("Verify() error = nil, want error when token is missing")
	}
}

func TestVerifySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	v := turnstile.NewVerifier(true, "secret", turnstile.WithEndpoint(server.URL))

	if err := v.Verify(context.Background(), "token", ""); err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
}

func TestVerifyFailureWithErrorCodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false,"error-codes":["invalid-input-response","timeout-or-duplicate"]}`))
	}))
	defer server.Close()

	v := turnstile.NewVerifier(true, "secret", turnstile.WithEndpoint(server.URL))

	err := v.Verify(context.Background(), "token", "")
	if err == nil {
		t.Fatal("Verify() error = nil, want failure")
	}
	if !strings.Contains(err.Error(), "invalid-input-response") {
		t.Fatalf("Verify() error = %v, want error codes included", err)
	}
}

func TestVerifyForwardsFieldsAndRemoteIP(t *testing.T) {
	var gotSecret, gotResponse, gotRemoteIP string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		gotSecret = r.FormValue("secret")
		gotResponse = r.FormValue("response")
		gotRemoteIP = r.FormValue("remoteip")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	v := turnstile.NewVerifier(true, "secret", turnstile.WithEndpoint(server.URL))

	if err := v.Verify(context.Background(), "token-123", "203.0.113.7"); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if gotSecret != "secret" {
		t.Fatalf("secret = %q, want %q", gotSecret, "secret")
	}
	if gotResponse != "token-123" {
		t.Fatalf("response = %q, want %q", gotResponse, "token-123")
	}
	if gotRemoteIP != "203.0.113.7" {
		t.Fatalf("remoteip = %q, want %q", gotRemoteIP, "203.0.113.7")
	}
}

func TestVerifyOmitsEmptyRemoteIP(t *testing.T) {
	var sawRemoteIP bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		_, sawRemoteIP = r.Form["remoteip"]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	v := turnstile.NewVerifier(true, "secret", turnstile.WithEndpoint(server.URL))

	if err := v.Verify(context.Background(), "token", ""); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if sawRemoteIP {
		t.Fatal("remoteip should be omitted when empty")
	}
}
