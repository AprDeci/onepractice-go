package handler

import (
	"io"
	"path/filepath"
	"strings"

	"onepractice-golang/internal/agent/llm"
	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"

	"github.com/gin-gonic/gin"
)

// AgentHandler 处理大模型相关的请求。
type AgentHandler struct {
	ocr *llm.GlmClient
}

func NewAgentHandler(ocr *llm.GlmClient) *AgentHandler {
	return &AgentHandler{ocr: ocr}
}

// OCR 图片文字识别接口。
// @Summary 图片 OCR 识别
// @Description 上传图片并调用 GLM OCR，返回识别出的文字与版面信息。
// @Tags agent
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param image formData file true "图片文件，支持 jpg/jpeg/png/webp，单张不超过 10MB"
// @Success 200 {object} response.Body{data=llm.OcrResult}
// @Failure 400 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/ocr [post]
func (h *AgentHandler) OCR(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "请上传图片"))
		return
	}

	// 限制大小（官方单图 ≤10MB）
	if file.Size > 10*1024*1024 {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "图片不能超过 10MB"))
		return
	}

	// 判断格式,只接受图片
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "只接受图片格式"))
		return
	}

	f, err := file.Open()
	if err != nil {
		response.Error(c, apperror.New(apperror.CodeInternal, err.Error()))
		return
	}
	defer f.Close()

	fileBytes, err := io.ReadAll(f)
	if err != nil {
		response.Error(c, apperror.New(apperror.CodeInternal, err.Error()))
		return
	}

	result, err := h.ocr.OCR(fileBytes, file.Filename)
	if err != nil {
		response.Error(c, apperror.New(apperror.CodeInternal, err.Error()))
		return
	}

	response.Success(c, result)
}
