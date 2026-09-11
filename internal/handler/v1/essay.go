package v1

import (
	"errors"

	"onepractice-golang/internal/agent"
	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	"onepractice-golang/internal/dto"
	dtoV1 "onepractice-golang/internal/dto/v1"
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
// @Description 提交作文题目与正文，创建异步批改任务，返回 taskId 与初始状态。
// @Tags essay
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateEssayTaskRequest true "作文批改任务参数"
// @Success 201 {object} response.Body{data=apiv1.EssayTaskCreatedResponse}
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

	taskID, err := h.service.CreateTask(c.Request.Context(), userID, req.RecordID, agent.Input{
		Title:   req.Title,
		Content: req.Content,
		Type:    req.Type,
	})
	if err != nil {
		response.Error(c, apperror.New(apperror.CodeInternal, err.Error()))
		return
	}

	response.Created(c, dtoV1.EssayTaskCreatedResponse{TaskID: taskID, Status: dto.EssayTaskPending})
}

// GetTask 查询作文批改任务。
// @Summary 查询作文批改任务
// @Description 按 taskId 查询当前登录用户的作文批改任务状态与结果。
// @Tags essay
// @Produce json
// @Security ApiKeyAuth
// @Param taskId path string true "任务 ID"
// @Success 200 {object} response.Body{data=dto.EssayTask}
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

// GetResultsByRecord 查询某次考试的作文评分结果。
// @Summary 查询考试的作文评分结果
// @Description 按 recordId 查询当前登录用户某次考试关联的作文评分结果列表。
// @Tags essay
// @Produce json
// @Security ApiKeyAuth
// @Param recordId path string true "答题记录 ID"
// @Success 200 {object} response.Body{data=[]dto.EssayResultResponse}
// @Router /api/v1/essay/records/{recordId}/results [get]
func (h *EssayHandler) GetResultsByRecord(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	recordID := c.Param("recordId")
	if recordID == "" {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "recordId 不能为空"))
		return
	}

	results, err := h.service.GetResultsByRecord(userID, recordID)
	if err != nil {
		if errors.Is(err, service.ErrDatabaseDisabled) {
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
			return
		}
		response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
		return
	}

	response.Success(c, results)
}
