package v1

import (
	"errors"
	"strconv"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	dto "onepractice-golang/internal/dto"
	dtoV1 "onepractice-golang/internal/dto/v1"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

type WordFavoriteHandler struct{ service *service.WordFavoriteService }

func NewWordFavoriteHandler(svc *service.WordFavoriteService) *WordFavoriteHandler {
	return &WordFavoriteHandler{service: svc}
}

// List 分页获取收藏单词。
// @Summary 分页获取收藏单词
// @Description 分页获取当前用户收藏的单词，并返回词库释义、音标等信息。
// @Tags word_favorite
// @Security ApiKeyAuth
// @Produce json
// @Param keyword query string false "关键词"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Body
// @Router /api/v1/users/me/favorite-words [get]
func (h *WordFavoriteHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var q dtoV1.FavoriteWordListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	q.PageQuery = q.PageQuery.Normalize()
	result, err := h.service.List(userID, dto.WordFavoriteListRequest{Keyword: q.Keyword, PageQuery: dto.PageQuery{Page: q.Page, PageSize: q.PageSize}})
	if err != nil {
		favoriteError(c, err)
		return
	}
	items := make([]dtoV1.CollectedWordItem, 0, len(result.Data))
	for _, item := range result.Data {
		items = append(items, dtoV1.CollectedWordItem{ID: item.ID, FavoriteID: item.FavoriteID, WordID: item.WordID, Word: item.Word, Spelling: item.Spelling, UKPhonetic: item.UKPhonetic, USPhonetic: item.USPhonetic, Paraphrase: item.Paraphrase, Frequency: item.Frequency, PaperID: item.PaperID, CreatedAt: item.CreatedAt})
	}
	response.Success(c, newPage(items, result.Total, q.PageQuery))
}

// Add 收藏单词。
// @Summary 收藏单词
// @Description 收藏词库中的单词。
// @Tags word_favorite
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body apiv1.FavoriteWordRequest true "收藏参数"
// @Success 201 {object} response.Body
// @Router /api/v1/users/me/favorite-words [post]
func (h *WordFavoriteHandler) Add(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var input dtoV1.FavoriteWordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	req := dto.WordFavoriteRequest{WordID: input.WordID, PaperID: input.PaperID}
	if err := h.service.Add(userID, req); err != nil {
		favoriteError(c, err)
		return
	}
	response.Created(c, req)
}

// Check 检查单词是否已收藏。
// @Summary 检查单词是否已收藏
// @Description 检查当前用户是否已收藏指定单词。
// @Tags word_favorite
// @Security ApiKeyAuth
// @Produce json
// @Param wordId path int true "单词 ID"
// @Success 200 {object} response.Body
// @Router /api/v1/users/me/favorite-words/{wordId} [get]
func (h *WordFavoriteHandler) Check(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	wordID, ok := pathUint(c, "wordId")
	if !ok {
		return
	}
	favorited, err := h.service.Has(userID, dto.WordFavoriteRequest{WordID: wordID})
	if err != nil {
		favoriteError(c, err)
		return
	}
	response.Success(c, dtoV1.FavoriteStatusResponse{WordID: wordID, Favorited: favorited})
}

// Remove 取消收藏单词。
// @Summary 取消收藏单词
// @Description 取消收藏词库中的单词。
// @Tags word_favorite
// @Security ApiKeyAuth
// @Param wordId path int true "单词 ID"
// @Success 204
// @Router /api/v1/users/me/favorite-words/{wordId} [delete]
func (h *WordFavoriteHandler) Remove(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	wordID, ok := pathUint(c, "wordId")
	if !ok {
		return
	}
	var paperID *int
	if value := c.Query("paperId"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
			return
		}
		paperID = &parsed
	}
	if err := h.service.Remove(userID, dto.WordFavoriteRequest{WordID: wordID, PaperID: paperID}); err != nil {
		favoriteError(c, err)
		return
	}
	response.NoContent(c)
}

func favoriteError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWordNotFound):
		response.Error(c, apperror.New(apperror.CodeNotFound, "word not found"))
	case errors.Is(err, service.ErrInvalidParam):
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
	case errors.Is(err, service.ErrDatabaseDisabled):
		response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
	default:
		response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
	}
}
