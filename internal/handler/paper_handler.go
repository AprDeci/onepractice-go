package handler

import (
	"net/http"
	"strconv"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/response"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

type PaperHandler struct {
	service *service.PaperService
}

func NewPaperHandler(service *service.PaperService) *PaperHandler {
	return &PaperHandler{service: service}
}

// All 获取全部试卷。
// @Summary 获取全部试卷
// @Description 返回全部试卷列表。
// @Tags paper
// @Produce json
// @Success 200 {object} response.Body
// @Router /api/paper/all [get]
func (h *PaperHandler) All(c *gin.Context) {
	papers, err := h.service.All()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, papers)
}

// Page 分页查询试卷。
// @Summary 分页查询试卷
// @Description 按试卷类型和年份分页查询试卷。
// @Tags paper
// @Accept json
// @Produce json
// @Param request body dto.PaperQueryRequest true "查询条件"
// @Success 200 {object} response.Body
// @Router /api/paper/getPaperwithQuerys [post]
func (h *PaperHandler) Page(c *gin.Context) {
	var req dto.PaperQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.Page(req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, result)
}

// PageWithRating 分页查询试卷和评分。
// @Summary 分页查询试卷和评分
// @Description 按试卷类型和年份分页查询试卷，同时返回评分和题目数量。
// @Tags paper
// @Accept json
// @Produce json
// @Param request body dto.PaperQueryRequest true "查询条件"
// @Success 200 {object} response.Body
// @Router /api/paper/getPaperandRatingWithQuerys [post]
func (h *PaperHandler) PageWithRating(c *gin.Context) {
	var req dto.PaperQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.PageWithRating(req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, result)
}

// ByType 按类型查询试卷。
// @Summary 按类型查询试卷
// @Description 根据试卷类型查询试卷列表。
// @Tags paper
// @Produce json
// @Param type query string true "试卷类型，如 CET-4/CET-6"
// @Success 200 {object} response.Body
// @Router /api/paper/type [get]
func (h *PaperHandler) ByType(c *gin.Context) {
	papers, err := h.service.ByType(c.Query("type"))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, papers)
}

// Types 获取全部试卷类型。
// @Summary 获取全部试卷类型
// @Description 返回数据库中存在的全部试卷类型。
// @Tags paper
// @Produce json
// @Success 200 {object} response.Body
// @Router /api/paper/types [get]
func (h *PaperHandler) Types(c *gin.Context) {
	types, err := h.service.Types()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, types)
}

// Intro 获取试卷简介。
// @Summary 获取试卷简介
// @Description 根据试卷 ID 返回试卷简介、分区数量和各分区题目数量。
// @Tags paper
// @Produce json
// @Param id query int true "试卷 ID"
// @Success 200 {object} response.Body
// @Router /api/paper/intro [get]
func (h *PaperHandler) Intro(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "id must be integer")
		return
	}

	intro, err := h.service.Intro(id)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, intro)
}
