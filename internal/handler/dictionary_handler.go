package handler

import (
	"errors"
	"net/http"
	"strconv"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/response"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DictionaryHandler struct {
	service *service.DictionaryService
}

func NewDictionaryHandler(service *service.DictionaryService) *DictionaryHandler {
	return &DictionaryHandler{service: service}
}

// ListWords 分页查询单词。
// @Summary 分页查询单词
// @Description 按关键词、拼写、释义、词书和词频范围分页查询单词。
// @Tags dictionary
// @Produce json
// @Param keyword query string false "在单词拼写和释义中搜索"
// @Param spelling query string false "按英文拼写模糊搜索"
// @Param paraphrase query string false "按中文释义模糊搜索"
// @Param bookid query int false "词书 ID"
// @Param min_frequency query number false "最小词频"
// @Param max_frequency query number false "最大词频"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页数量，默认 20，最大 100"
// @Success 200 {object} response.Body
// @Router /api/dictionary/words [get]
func (h *DictionaryHandler) ListWords(c *gin.Context) {
	var req dto.DictionaryWordListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.ListWords(req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, result)
}

// LookupMeanings 按单词查询释义。
// @Summary 按单词查询释义
// @Description 根据英文拼写查询中文释义，可选择精确匹配或模糊匹配。
// @Tags dictionary
// @Produce json
// @Param spelling query string true "英文拼写"
// @Param exact query bool false "是否精确匹配，默认 false"
// @Param limit query int false "返回数量，默认 20，最大 100"
// @Success 200 {object} response.Body
// @Router /api/dictionary/lookup [get]
func (h *DictionaryHandler) LookupMeanings(c *gin.Context) {
	var req dto.DictionaryLookupRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.LookupMeanings(req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, result)
}

// GetWordDetail 获取单词详情。
// @Summary 获取单词详情
// @Description 根据单词 ID 获取单词基础信息、所属词书和例句。
// @Tags dictionary
// @Produce json
// @Param wordid path int true "单词 ID"
// @Success 200 {object} response.Body
// @Router /api/dictionary/words/{wordid} [get]
func (h *DictionaryHandler) GetWordDetail(c *gin.Context) {
	wordID, err := strconv.ParseUint(c.Param("wordid"), 10, 64)
	if err != nil || wordID == 0 {
		response.Error(c, http.StatusBadRequest, "wordid must be positive integer")
		return
	}

	result, err := h.service.GetWordDetail(uint(wordID))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "word not found")
		return
	}
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, result)
}

// GetWordBySpelling 按拼写获取单词详情。
// @Summary 按拼写获取单词详情
// @Description 根据英文拼写精确获取单词基础信息、所属词书和例句。
// @Tags dictionary
// @Produce json
// @Param spelling path string true "英文拼写"
// @Success 200 {object} response.Body
// @Router /api/dictionary/words/spelling/{spelling} [get]
func (h *DictionaryHandler) GetWordBySpelling(c *gin.Context) {
	result, err := h.service.GetWordBySpelling(c.Param("spelling"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "word not found")
		return
	}
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, result)
}

// ListBooks 分页查询词书。
// @Summary 分页查询词书
// @Description 按词书名称和状态分页查询词书。
// @Tags dictionary
// @Produce json
// @Param keyword query string false "词书名称关键词"
// @Param status query int false "词书状态"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页数量，默认 20，最大 100"
// @Success 200 {object} response.Body
// @Router /api/dictionary/books [get]
func (h *DictionaryHandler) ListBooks(c *gin.Context) {
	var req dto.DictionaryBookListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.ListBooks(req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, result)
}

// ListBookWords 分页查询词书单词。
// @Summary 分页查询词书单词
// @Description 根据词书 ID 分页查询该词书下的单词，支持关键词搜索。
// @Tags dictionary
// @Produce json
// @Param bookid path int true "词书 ID"
// @Param keyword query string false "在单词拼写和释义中搜索"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页数量，默认 20，最大 100"
// @Success 200 {object} response.Body
// @Router /api/dictionary/books/{bookid}/words [get]
func (h *DictionaryHandler) ListBookWords(c *gin.Context) {
	bookID, err := strconv.ParseUint(c.Param("bookid"), 10, 64)
	if err != nil || bookID == 0 {
		response.Error(c, http.StatusBadRequest, "bookid must be positive integer")
		return
	}

	var req dto.DictionaryBookWordsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.ListBookWords(uint(bookID), req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, result)
}
