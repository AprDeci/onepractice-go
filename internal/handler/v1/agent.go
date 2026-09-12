package v1

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"onepractice-golang/internal/agent/llm"
	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	dtoV1 "onepractice-golang/internal/dto/v1"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AgentHandler 处理大模型相关的请求。
type AgentHandler struct {
	ocr    *llm.GlmClient
	points *service.PointsService
}

func NewAgentHandler(ocr *llm.GlmClient, points *service.PointsService) *AgentHandler {
	return &AgentHandler{ocr: ocr, points: points}
}

// OCR 图片文字识别接口。
// @Summary 图片 OCR 识别
// @Description 上传图片并调用 GLM OCR，返回识别出的文字与版面信息。
// @Tags agent
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param image formData file true "图片文件，支持 jpg/jpeg/png/webp，单张不超过 10MB"
// @Success 200 {object} response.Body{data=apiv1.OcrResponse}
// @Failure 400 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/v1/ocr [post]
func (h *AgentHandler) OCR(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

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

	requestID := ocrRequestID(c)
	if h.points != nil {
		cost, err := h.points.CostOf(c.Request.Context(), service.PointActionOCR)
		if err != nil {
			response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
			return
		}
		if err := h.points.Deduct(c.Request.Context(), userID, cost, service.PointTypeOCRSpend, requestID, "OCR识别"); err != nil {
			switch {
			case errors.Is(err, service.ErrInsufficientPoints):
				response.Error(c, apperror.New(apperror.CodeInsufficientPoints, "积分不足"))
			case errors.Is(err, service.ErrDatabaseDisabled):
				response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
			default:
				response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
			}
			return
		}
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
		if h.points != nil {
			_ = h.points.Refund(context.WithoutCancel(c.Request.Context()), userID, service.PointTypeOCRSpend, requestID, "OCR识别失败")
		}
		response.Error(c, apperror.New(apperror.CodeInternal, err.Error()))
		return
	}

	response.Success(c, dtoV1.OcrResponse{MdResult: result.MdResult, LayoutDetail: result.LayoutDetail})
}

// ocrRequestID 取请求链路 ID 作为扣费幂等键，缺失时兜底生成。
func ocrRequestID(c *gin.Context) string {
	if v, ok := c.Get(response.RequestIDKey); ok {
		if id, _ := v.(string); id != "" {
			return id
		}
	}
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
