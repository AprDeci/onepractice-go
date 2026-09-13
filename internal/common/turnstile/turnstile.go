package turnstile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// defaultVerifyURL 是 Cloudflare Turnstile 的官方校验端点。
const defaultVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// Verifier 校验 Cloudflare Turnstile token；enabled=false 时恒跳过校验，
// 便于在未配置密钥的环境下安全部署。
type Verifier struct {
	enabled   bool
	secretKey string
	client    *http.Client
	verifyURL string
}

// Option 配置 Verifier 的可选参数。
type Option func(*Verifier)

// WithEndpoint 覆盖 siteverify 端点，主要用于测试注入本地 mock。
func WithEndpoint(endpoint string) Option {
	return func(v *Verifier) {
		if endpoint != "" {
			v.verifyURL = endpoint
		}
	}
}

// NewVerifier 创建校验器；enabled=false 时 Verify 恒返回 nil（跳过校验）。
func NewVerifier(enabled bool, secretKey string, opts ...Option) *Verifier {
	v := &Verifier{
		enabled:   enabled,
		secretKey: secretKey,
		client:    &http.Client{Timeout: 5 * time.Second},
		verifyURL: defaultVerifyURL,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// siteverifyResponse 对应 Cloudflare siteverify 的响应体。
type siteverifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

// Verify 校验一次性 Turnstile token。enabled=true 且 secretKey 为空视为配置错误。
// 错误信息与日志均不包含 token 或 secret。
func (v *Verifier) Verify(ctx context.Context, token, remoteIP string) error {
	if v == nil || !v.enabled {
		return nil
	}
	if strings.TrimSpace(v.secretKey) == "" {
		return errors.New("turnstile secret key is not configured")
	}
	if strings.TrimSpace(token) == "" {
		return errors.New("turnstile token is empty")
	}

	form := url.Values{}
	form.Set("secret", v.secretKey)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build turnstile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("verify turnstile: %w", err)
	}
	defer resp.Body.Close()

	var result siteverifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode turnstile response: %w", err)
	}
	if !result.Success {
		if len(result.ErrorCodes) > 0 {
			return fmt.Errorf("turnstile verification failed: %s", strings.Join(result.ErrorCodes, ","))
		}
		return errors.New("turnstile verification failed")
	}
	return nil
}
