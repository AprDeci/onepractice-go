package handler

import (
	"strconv"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/response"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

type RecordHandler struct {
	service *service.RecordService
}

func NewRecordHandler(service *service.RecordService) *RecordHandler {
	return &RecordHandler{service: service}
}

// Save 保存答题记录。
// @Summary 保存答题记录
// @Description 为当前登录用户创建一条答题记录，返回 recordId。
// @Tags record
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.RecordRequest true "答题记录参数"
// @Success 200 {object} response.Body
// @Router /api/record/save [post]
func (h *RecordHandler) Save(c *gin.Context) {
	userID, ok := loginID(c)
	if !ok {
		return
	}
	var req dto.RecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorEnum(c, response.ErrParamInvalid)
		return
	}
	recordID, err := h.service.Create(userID, req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, recordID)
}

// List 获取答题记录列表。
// @Summary 获取答题记录列表
// @Description 按最近天数分页获取当前登录用户的答题记录。
// @Tags record
// @Produce json
// @Security ApiKeyAuth
// @Param days query int false "最近天数"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Body
// @Router /api/record/list [get]
func (h *RecordHandler) List(c *gin.Context) {
	userID, ok := loginID(c)
	if !ok {
		return
	}
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	records, err := h.service.ListRecent(userID, days, pageNum, pageSize)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, records)
}

// Update 更新答题记录。
// @Summary 更新答题记录
// @Description 更新当前登录用户已有答题记录的作答结果。
// @Tags record
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.RecordRequest true "答题记录参数"
// @Success 200 {object} response.Body
// @Router /api/record/update [post]
func (h *RecordHandler) Update(c *gin.Context) {
	userID, ok := loginID(c)
	if !ok {
		return
	}
	var req dto.RecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorEnum(c, response.ErrParamInvalid)
		return
	}
	if err := h.service.Update(userID, req); err != nil {
		writeError(c, err)
		return
	}
	response.SuccessNoData(c)
}
