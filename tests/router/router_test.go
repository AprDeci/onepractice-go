package router_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
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
	engine, _, cleanup, err := router.New(cfg, nil, nil, logger)
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
		{http.MethodGet, "/api/v1/points/costs"},
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
		{http.MethodPatch, "/api/v1/users/me"},
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
		{http.MethodGet, "/api/v1/essay/records/abc/results"},
		{http.MethodPost, "/api/v1/ocr"},
		{http.MethodGet, "/api/v1/points/balance"},
		{http.MethodGet, "/api/v1/points/transactions"},
		{http.MethodPost, "/api/v1/points/daily-checkin"},
		{http.MethodGet, "/api/v1/points/checkin-status"},
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

// /api 下的旧路由已全部移除，任何残留都会在 401/200 而不是 404 上暴露出来。
func TestLegacyRoutesRemoved(t *testing.T) {
	r := newTestEngine(t)
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/user/register"},
		{http.MethodPost, "/api/user/login"},
		{http.MethodPost, "/api/user/resetpassword"},
		{http.MethodGet, "/api/captcha/email"},
		{http.MethodPost, "/api/captcha/email/verify"},
		{http.MethodGet, "/api/paper/all"},
		{http.MethodPost, "/api/paper/getPaperwithQuerys"},
		{http.MethodGet, "/api/paper/types"},
		{http.MethodGet, "/api/question/getById"},
		{http.MethodGet, "/api/dictionary/words"},
		{http.MethodGet, "/api/dictionary/books/1/words"},
		{http.MethodGet, "/api/user/info"},
		{http.MethodPost, "/api/record/save"},
		{http.MethodPost, "/api/word/favorites"},
		{http.MethodPost, "/api/ocr"},
	}
	for _, tc := range cases {
		w := doRequest(t, r, tc.method, tc.path)
		if w.Code != http.StatusNotFound {
			t.Errorf("%s %s: want 404 (legacy route removed), got %d", tc.method, tc.path, w.Code)
		}
	}
}

func TestHealthHasNoDeprecationHeader(t *testing.T) {
	r := newTestEngine(t)
	w := doRequest(t, r, http.MethodGet, "/health")
	if got := w.Header().Get("Deprecation"); got != "" {
		t.Errorf("/health must not carry Deprecation header, got %q", got)
	}
}

// 依赖不可用（nil DB / nil Redis）必须暴露成 503「依赖服务不可用」，
// 不能被 handler 的 default 分支吞成 500「系统异常」。
func TestDependencyUnavailableSurfacesAsServiceUnavailable(t *testing.T) {
	r := newTestEngine(t)
	token, err := sagin.Login(int64(1))
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	t.Run("essayGetTaskRedisDisabled", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/essay/tasks/abc", nil)
		req.Header.Set("satoken", token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		requireServiceUnavailable(t, w)
	})

	t.Run("ocrCostOfDatabaseDisabled", func(t *testing.T) {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		part, err := mw.CreateFormFile("image", "test.jpg")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write([]byte("not-a-real-jpeg")); err != nil {
			t.Fatalf("write form file: %v", err)
		}
		if err := mw.Close(); err != nil {
			t.Fatalf("close multipart writer: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/ocr", &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("satoken", token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		requireServiceUnavailable(t, w)
	})
}

func requireServiceUnavailable(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 (body: %s)", w.Code, w.Body.String())
	}
	var body response.Body
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body %q: %v", w.Body.String(), err)
	}
	if body.Code != apperror.CodeServiceUnavailable || body.Message != "依赖服务不可用" {
		t.Fatalf("body = {code:%d message:%q}, want {code:%d message:%q}",
			body.Code, body.Message, apperror.CodeServiceUnavailable, "依赖服务不可用")
	}
}
