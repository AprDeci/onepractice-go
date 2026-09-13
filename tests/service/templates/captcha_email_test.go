package templates_test

import (
	"strings"
	"testing"

	"onepractice-golang/internal/service/templates"
)

func TestRenderCaptchaEmailRegister(t *testing.T) {
	body, err := templates.RenderCaptchaEmail(templates.CaptchaEmailData{
		Code:          "123456",
		Intro:         "你正在进行注册",
		ExpireMinutes: 5,
	})
	if err != nil {
		t.Fatalf("RenderCaptchaEmail() error = %v", err)
	}

	for _, want := range []string{"123456", "你正在进行注册", "5 分钟"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
	for _, unwanted := range []string{"%s", "{{"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("body contains unreplaced placeholder %q", unwanted)
		}
	}
}

func TestRenderCaptchaEmailResetPassword(t *testing.T) {
	body, err := templates.RenderCaptchaEmail(templates.CaptchaEmailData{
		Code:          "654321",
		Intro:         "你正在进行密码重置",
		ExpireMinutes: 5,
	})
	if err != nil {
		t.Fatalf("RenderCaptchaEmail() error = %v", err)
	}

	for _, want := range []string{"654321", "你正在进行密码重置", "5 分钟"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}
