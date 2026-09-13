package middleware_test

import (
	"context"
	"net/http/httptest"
	"onepractice-golang/internal/middleware"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestTimeoutAddsDeadlineToRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.TimeoutMiddleware(50 * time.Millisecond))
	router.GET("/test", func(c *gin.Context) {
		deadline, ok := c.Request.Context().Deadline()
		if !ok {
			t.Fatal("request context has no deadline")
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			t.Fatalf("deadline already expired: %v", remaining)
		}
		c.Status(204)
	})

	request := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != 204 {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
}

func TestTimeoutCancelsRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.TimeoutMiddleware(10 * time.Millisecond))
	router.GET("/test", func(c *gin.Context) {
		select {
		case <-time.After(100 * time.Millisecond):
			t.Fatal("request context was not canceled")

		case <-c.Request.Context().Done():
			if err := c.Request.Context().Err(); err != context.DeadlineExceeded {
				t.Fatalf(
					"context error = %v, want %v",
					err,
					context.DeadlineExceeded,
				)
			}
			c.Status(204)
		}
	})

	request := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != 204 {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
}
