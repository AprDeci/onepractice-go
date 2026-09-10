package v1

import (
	"errors"

	"onepractice-golang/internal/agent"
	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

// EssayHandler 处理作文批改任务接口。
type EssayHandler struct {
	service *service.EssayService
}

func NewEssayHandler(s *service.EssayService) *EssayHandler {
	return &EssayHandler{service: s}
}

// CreateTask 创建作文批改任务。
// @Summary 创建作文批改任务
// @Tags essay
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 201 {object} response.Body
// @Router /api/v1/essay/tasks [post]
func (h *EssayHandler) CreateTask(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var req dto.CreateEssayTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数错误"))
		return
	}
	if req.Type != "四级" && req.Type != "六级" {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "type 只能是四级或六级"))
		return
	}

	taskID, err := h.service.CreateTask(c.Request.Context(), userID, agent.Input{
		Title:   req.Title,
		Content: req.Content,
		Type:    req.Type,
	})
	if err != nil {
		response.Error(c, apperror.New(apperror.CodeInternal, err.Error()))
		return
	}

	response.Created(c, gin.H{"taskId": taskID, "status": dto.EssayTaskPending})
}

// GetTask 查询作文批改任务。
// @Summary 查询作文批改任务
// @Tags essay
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Body
// @Router /api/v1/essay/tasks/{taskId} [get]
func (h *EssayHandler) GetTask(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	task, err := h.service.GetTask(c.Request.Context(), userID, c.Param("taskId"))
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			response.Error(c, apperror.New(apperror.CodeNotFound, "任务不存在"))
			return
		}
		response.Error(c, apperror.New(apperror.CodeInternal, err.Error()))
		return
	}

	response.Success(c, task)
}
