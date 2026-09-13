package templates

import (
	"bytes"
	_ "embed"
	"html/template"
)

//go:embed captcha_code.html
var captchaEmailHTML string

var captchaEmailTmpl = template.Must(template.New("captcha_code").Parse(captchaEmailHTML))

// CaptchaEmailData 是验证码邮件的渲染数据。
type CaptchaEmailData struct {
	Code          string
	Intro         string
	ExpireMinutes int
}

// RenderCaptchaEmail 渲染验证码邮件 HTML 正文。
func RenderCaptchaEmail(data CaptchaEmailData) (string, error) {
	var buf bytes.Buffer
	if err := captchaEmailTmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
