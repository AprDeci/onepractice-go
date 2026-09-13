package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"onepractice-golang/internal/common/turnstile"
	"onepractice-golang/internal/middleware"

	"github.com/gin-gonic/gin"
)

func newTurnstileEngine(v *turnstile.Verifier) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/x", middleware.VerifyTurnstile(v), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func doTurnstileRequest(r *gin.Engine, token string) int {
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	if token != "" {
		req.Header.Set(middleware.TurnstileTokenHeader, token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func newSiteverifyServer(t *testing.T, body string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestTurnstileTokenHeaderConstant(t *testing.T) {
	if middleware.TurnstileTokenHeader != "X-Turnstile-Token" {
		t.Fatalf("TurnstileTokenHeader = %q, want %q", middleware.TurnstileTokenHeader, "X-Turnstile-Token")
	}
}

func TestVerifyTurnstileNilVerifierPasses(t *testing.T) {
	r := newTurnstileEngine(nil)

	if code := doTurnstileRequest(r, ""); code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for nil verifier", code)
	}
}

func TestVerifyTurnstileDisabledPassesWithoutHeader(t *testing.T) {
	r := newTurnstileEngine(turnstile.NewVerifier(false, ""))

	if code := doTurnstileRequest(r, ""); code != http.StatusOK {
		t.Fatalf("status = %d, want 200 when disabled", code)
	}
}

func TestVerifyTurnstileEnabledMissingHeaderRejected(t *testing.T) {
	server := newSiteverifyServer(t, `{"success":true}`)
	r := newTurnstileEngine(turnstile.NewVerifier(true, "secret", turnstile.WithEndpoint(server.URL)))

	if code := doTurnstileRequest(r, ""); code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 when header missing", code)
	}
}

func TestVerifyTurnstileEnabledSuccessPasses(t *testing.T) {
	server := newSiteverifyServer(t, `{"success":true}`)
	r := newTurnstileEngine(turnstile.NewVerifier(true, "secret", turnstile.WithEndpoint(server.URL)))

	if code := doTurnstileRequest(r, "valid-token"); code != http.StatusOK {
		t.Fatalf("status = %d, want 200 on success", code)
	}
}

func TestVerifyTurnstileEnabledFailureRejected(t *testing.T) {
	server := newSiteverifyServer(t, `{"success":false,"error-codes":["invalid-input-response"]}`)
	r := newTurnstileEngine(turnstile.NewVerifier(true, "secret", turnstile.WithEndpoint(server.URL)))

	if code := doTurnstileRequest(r, "bad-token"); code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 on failure", code)
	}
}
