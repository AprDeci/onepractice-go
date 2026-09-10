package router_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"onepractice-golang/internal/config"
	"onepractice-golang/internal/router"

	"github.com/gin-gonic/gin"
	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
	"github.com/sa-tokens/sa-token-go/storage/memory"
)

var initAuthOnce sync.Once

func initAuthManager() {
	initAuthOnce.Do(func() {
		saConfig := sagin.DefaultConfig()
		saConfig.IsPrintBanner = false
		sagin.SetManager(sagin.NewManager(memory.NewStorage(), saConfig))
	})
}

func newTestEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	initAuthManager()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{LLM: config.LLMConfig{
		Default: "test",
		Models: map[string]config.LLMModelConfig{
			"test": {BaseURL: "https://example.com/v1", Model: "test-model", APIKey: "test-key"},
		},
	}}
	engine, cleanup, err := router.New(cfg, nil, nil, logger)
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	t.Cleanup(cleanup)
	return engine
}

func doRequest(t *testing.T, r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestV1PublicRoutesRegistered(t *testing.T) {
	r := newTestEngine(t)
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/auth/registrations"},
		{http.MethodPost, "/api/v1/auth/sessions"},
		{http.MethodPost, "/api/v1/auth/password-resets"},
		{http.MethodPost, "/api/v1/auth/email-verifications"},
		{http.MethodPost, "/api/v1/auth/email-verifications/verification"},
		{http.MethodGet, "/api/v1/papers"},
		{http.MethodGet, "/api/v1/papers/1/intro"},
		{http.MethodGet, "/api/v1/paper-types"},
		{http.MethodGet, "/api/v1/papers/1/questions"},
		{http.MethodGet, "/api/v1/papers/1/answers"},
		{http.MethodPost, "/api/v1/questions/practice"},
		{http.MethodGet, "/api/v1/dictionary/definitions"},
		{http.MethodGet, "/api/v1/dictionary/words"},
		{http.MethodGet, "/api/v1/dictionary/words/1"},
		{http.MethodGet, "/api/v1/dictionary/words/by-spelling/test"},
		{http.MethodGet, "/api/v1/dictionary/books"},
		{http.MethodGet, "/api/v1/dictionary/books/1/words"},
	}
	for _, tc := range cases {
		w := doRequest(t, r, tc.method, tc.path)
		if w.Code == http.StatusNotFound {
			t.Errorf("%s %s: route not registered (404)", tc.method, tc.path)
		}
	}
}

func TestV1ProtectedRoutesRequireAuth(t *testing.T) {
	r := newTestEngine(t)
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/users/me"},
		{http.MethodDelete, "/api/v1/auth/sessions/current"},
		{http.MethodPost, "/api/v1/records"},
		{http.MethodGet, "/api/v1/records"},
		{http.MethodPut, "/api/v1/records/abc"},
		{http.MethodGet, "/api/v1/users/me/favorite-words"},
		{http.MethodPost, "/api/v1/users/me/favorite-words"},
		{http.MethodGet, "/api/v1/users/me/favorite-words/1"},
		{http.MethodDelete, "/api/v1/users/me/favorite-words/1"},
		{http.MethodPost, "/api/v1/essay/tasks"},
		{http.MethodGet, "/api/v1/essay/tasks/abc"},
	}
	for _, tc := range cases {
		w := doRequest(t, r, tc.method, tc.path)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: want 401, got %d", tc.method, tc.path, w.Code)
		}
	}
}

func TestLegacyRoutesRegisteredWithDeprecationHeaders(t *testing.T) {
	r := newTestEngine(t)
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/user/login"},
		{http.MethodGet, "/api/paper/all"},
		{http.MethodPost, "/api/paper/getPaperwithQuerys"},
		{http.MethodGet, "/api/paper/types"},
		{http.MethodGet, "/api/question/getById"},
		{http.MethodGet, "/api/dictionary/words"},
		{http.MethodGet, "/api/dictionary/words/1"},
		{http.MethodGet, "/api/dictionary/books/1/words"},
	}
	for _, tc := range cases {
		w := doRequest(t, r, tc.method, tc.path)
		if w.Code == http.StatusNotFound {
			t.Errorf("legacy %s %s: route not registered (404)", tc.method, tc.path)
			continue
		}
		if got := w.Header().Get("Deprecation"); got != "true" {
			t.Errorf("legacy %s %s: Deprecation=%q, want \"true\"", tc.method, tc.path, got)
		}
		if got := w.Header().Get("Sunset"); got == "" {
			t.Errorf("legacy %s %s: missing Sunset header", tc.method, tc.path)
		}
	}
}

func TestLegacyProtectedRoutesRequireAuth(t *testing.T) {
	r := newTestEngine(t)
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/user/info"},
		{http.MethodPost, "/api/record/save"},
		{http.MethodGet, "/api/record/list"},
		{http.MethodPost, "/api/word/favorites"},
		{http.MethodGet, "/api/word/favorites"},
		{http.MethodPost, "/api/ocr"},
	}
	for _, tc := range cases {
		w := doRequest(t, r, tc.method, tc.path)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: want 401, got %d", tc.method, tc.path, w.Code)
		}
	}
}

func TestV1RoutesHaveNoDeprecationHeaders(t *testing.T) {
	r := newTestEngine(t)
	for _, path := range []string{"/api/v1/paper-types", "/api/v1/dictionary/words"} {
		w := doRequest(t, r, http.MethodGet, path)
		if w.Code == http.StatusNotFound {
			t.Errorf("%s: route not registered", path)
			continue
		}
		if got := w.Header().Get("Deprecation"); got != "" {
			t.Errorf("%s: must not carry Deprecation header, got %q", path, got)
		}
		if got := w.Header().Get("Sunset"); got != "" {
			t.Errorf("%s: must not carry Sunset header, got %q", path, got)
		}
	}
}

func TestV1InvalidPathParameterReturns400(t *testing.T) {
	r := newTestEngine(t)
	w := doRequest(t, r, http.MethodGet, "/api/v1/papers/not-an-int/intro")
	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400 for invalid paperId, got %d", w.Code)
	}
}

func TestV1OCRIsNotRegistered(t *testing.T) {
	r := newTestEngine(t)
	w := doRequest(t, r, http.MethodPost, "/api/v1/ocr")
	if w.Code != http.StatusNotFound {
		t.Errorf("POST /api/v1/ocr: want 404 (OCR is legacy-only), got %d", w.Code)
	}
}

func TestHealthHasNoDeprecationHeader(t *testing.T) {
	r := newTestEngine(t)
	w := doRequest(t, r, http.MethodGet, "/health")
	if got := w.Header().Get("Deprecation"); got != "" {
		t.Errorf("/health must not carry Deprecation header, got %q", got)
	}
}
