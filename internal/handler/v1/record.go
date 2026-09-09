package v1

import (
	"errors"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	dto "onepractice-golang/internal/dto"
	dtoV1 "onepractice-golang/internal/dto/v1"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

type RecordHandler struct{ service *service.RecordService }

func NewRecordHandler(svc *service.RecordService) *RecordHandler { return &RecordHandler{service: svc} }

func recordRequest(recordID string, req dtoV1.CreateRecordRequest) dto.RecordRequest {
	return dto.RecordRequest{RecordID: recordID, PaperID: dto.StringInt(req.PaperID), Type: req.Type, IsFinished: req.IsFinished, Answers: req.Answers, Score: req.Score, TotalScore: req.TotalScore, TimeSpend: req.TimeSpend, HasSpendTime: dto.StringInt64(req.HasSpendTime)}
}

// Create 保存答题记录。
// @Summary 保存答题记录
// @Description 为当前登录用户创建一条答题记录，返回 recordId。
// @Tags record
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body apiv1.CreateRecordRequest true "答题记录参数"
// @Success 201 {object} response.Body
// @Router /api/v1/records [post]
func (h *RecordHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var input dtoV1.CreateRecordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	recordID, err := h.service.Create(userID, recordRequest("", input))
	if err != nil {
		recordError(c, err)
		return
	}
	response.Created(c, dtoV1.RecordCreatedResponse{RecordID: recordID})
}

// List 获取答题记录列表。
// @Summary 获取答题记录列表
// @Description 按最近天数分页获取当前登录用户的答题记录。
// @Tags record
// @Security ApiKeyAuth
// @Produce json
// @Param days query int false "最近天数"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Body
// @Router /api/v1/records [get]
func (h *RecordHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req dtoV1.RecordListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	records, total, err := h.service.ListRecentPage(userID, req.Days, req.Page, req.PageSize)
	if err != nil {
		recordListError(c, err)
		return
	}
	items := make([]dtoV1.UserExamRecord, 0, len(records))
	for _, record := range records {
		items = append(items, dtoV1.UserExamRecord{RecordID: record.RecordID, UserID: record.UserID, PaperID: record.PaperID, PaperType: record.PaperType, PaperName: record.PaperName, Type: record.Type, IsFinished: record.IsFinished, Answers: record.Answers, TimeSpend: record.TimeSpend, Score: record.Score, TotalScore: record.TotalScore, Timestamp: record.Timestamp, HasSpendTime: record.HasSpendTime})
	}
	response.Success(c, newPage(items, total, req.PageQuery))
}

// Update 更新答题记录。
// @Summary 更新答题记录
// @Description 更新当前登录用户已有答题记录的作答结果。
// @Tags record
// @Security ApiKeyAuth
// @Accept json
// @Param recordId path string true "记录 ID"
// @Param request body apiv1.UpdateRecordRequest true "答题记录参数"
// @Success 204
// @Router /api/v1/records/{recordId} [put]
func (h *RecordHandler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	recordID := c.Param("recordId")
	if recordID == "" {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	var input dtoV1.UpdateRecordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	if err := h.service.Update(userID, recordRequest(recordID, dtoV1.CreateRecordRequest{PaperID: input.PaperID, Type: input.Type, IsFinished: input.IsFinished, Answers: input.Answers, Score: input.Score, TotalScore: input.TotalScore, TimeSpend: input.TimeSpend, HasSpendTime: input.HasSpendTime})); err != nil {
		updateRecordError(c, err)
		return
	}
	response.NoContent(c)
}

func recordError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidParam):
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
	case errors.Is(err, service.ErrTokenInvalid):
		response.Error(c, apperror.New(apperror.CodeUnauthorized, "Token失效"))
	case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
		response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
	default:
		response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
	}
}
func recordListError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidParam):
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
	case errors.Is(err, service.ErrRedisDisabled):
		response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
	default:
		response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
	}
}
func updateRecordError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidParam):
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
	case errors.Is(err, service.ErrRedisDisabled):
		response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
	default:
		response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
	}
}
