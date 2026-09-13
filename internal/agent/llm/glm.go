package llm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// GlmClient 封装智谱 GLM 开放平台调用，持有 API Key 与 HTTP 客户端。
type GlmClient struct {
	apiKey string
	client *http.Client
}

type OcrResult struct {
	MdResult     string                     `json:"md_results"`
	LayoutDetail [][]map[string]interface{} `json:"layout_details"`
}

func NewGlmClient(apiKey string) *GlmClient {
	return &GlmClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// OCR 识别图片中的文字与版面信息。
func (c *GlmClient) OCR(fileBytes []byte, fileName string) (OcrResult, error) {
	if c == nil || strings.TrimSpace(c.apiKey) == "" {
		return OcrResult{}, fmt.Errorf("glm api key 未配置")
	}

	// 1. 判断 mime
	ext := strings.ToLower(filepath.Ext(fileName))
	var mime string
	switch ext {
	case ".jpg", ".jpeg":
		mime = "image/jpeg"
	case ".png":
		mime = "image/png"
	case ".webp":
		mime = "image/webp"
	default:
		mime = "image/png" // 兜底
	}

	// 2. 转成 data URI（官方推荐格式）
	b64 := base64.StdEncoding.EncodeToString(fileBytes)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mime, b64)

	// 3. 构造请求体
	body := map[string]any{
		"model": "glm-ocr",
		"file":  dataURI,
		// 可选参数
		// "return_crop_images": true,
		// "need_layout_visualization": true,
	}

	jsonData, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, "https://open.bigmodel.cn/api/paas/v4/layout_parsing", bytes.NewReader(jsonData))
	if err != nil {
		return OcrResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return OcrResult{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return OcrResult{}, fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	var result OcrResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return OcrResult{}, err
	}

	return result, nil
}
