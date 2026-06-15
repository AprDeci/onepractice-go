package handler

import (
	"errors"
	"net/http"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/response"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

type WordFavoriteHandler struct {
	service *service.WordFavoriteService
}

func NewWordFavoriteHandler(service *service.WordFavoriteService) *WordFavoriteHandler {
	return &WordFavoriteHandler{service: service}
}

// Add 收藏单词。
// @Summary 收藏单词
// @Description 收藏词库中的单词。优先使用 wordid；未传 wordid 时按 word 精确匹配 tb_vocabulary.spelling。
// @Tags word
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.WordFavoriteRequest true "收藏参数"
// @Success 200 {object} response.Body
// @Router /api/word/favorites [post]
func (h *WordFavoriteHandler) Add(c *gin.Context) {
	userID, ok := loginID(c)
	if !ok {
		return
	}

	var req dto.WordFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Add(userID, req); err != nil {
		h.writeWordFavoriteError(c, err)
		return
	}
	response.SuccessNoData(c)
}

// Remove 取消收藏单词。
// @Summary 取消收藏单词
// @Description 取消收藏词库中的单词。支持 wordid 或 word 查询参数；不存在也返回成功。
// @Tags word
// @Produce json
// @Security ApiKeyAuth
// @Param wordid query int false "单词 ID"
// @Param word query string false "英文拼写"
// @Success 200 {object} response.Body
// @Router /api/word/favorites [delete]
func (h *WordFavoriteHandler) Remove(c *gin.Context) {
	userID, ok := loginID(c)
	if !ok {
		return
	}

	var req dto.WordFavoriteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Remove(userID, req); err != nil {
		h.writeWordFavoriteError(c, err)
		return
	}
	response.SuccessNoData(c)
}

// Check 检查单词是否已收藏。
// @Summary 检查单词是否已收藏
// @Description 检查当前用户是否已收藏指定单词。支持 wordid 或 word 查询参数。
// @Tags word
// @Produce json
// @Security ApiKeyAuth
// @Param wordid query int false "单词 ID"
// @Param word query string false "英文拼写"
// @Success 200 {object} response.Body
// @Router /api/word/favorites/check [get]
func (h *WordFavoriteHandler) Check(c *gin.Context) {
	userID, ok := loginID(c)
	if !ok {
		return
	}

	var req dto.WordFavoriteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	has, err := h.service.Has(userID, req)
	if err != nil {
		h.writeWordFavoriteError(c, err)
		return
	}
	response.Success(c, has)
}

// List 分页获取收藏单词。
// @Summary 分页获取收藏单词
// @Description 分页获取当前用户收藏的单词，并返回词库释义、音标等信息。
// @Tags word
// @Produce json
// @Security ApiKeyAuth
// @Param keyword query string false "在单词拼写和释义中搜索"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页数量，默认 20，最大 100"
// @Success 200 {object} response.Body
// @Router /api/word/favorites [get]
func (h *WordFavoriteHandler) List(c *gin.Context) {
	userID, ok := loginID(c)
	if !ok {
		return
	}

	var req dto.WordFavoriteListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.List(userID, req)
	if err != nil {
		h.writeWordFavoriteError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *WordFavoriteHandler) writeWordFavoriteError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrWordNotFound) {
		response.Error(c, http.StatusNotFound, "word not found")
		return
	}
	writeError(c, err)
}
