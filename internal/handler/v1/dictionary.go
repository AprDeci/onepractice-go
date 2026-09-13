package v1

import (
	"errors"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	dto "onepractice-golang/internal/dto"
	dtoV1 "onepractice-golang/internal/dto/v1"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DictionaryHandler struct {
	service *service.DictionaryService
}

func NewDictionaryHandler(svc *service.DictionaryService) *DictionaryHandler {
	return &DictionaryHandler{service: svc}
}

// Lookup 按单词查询释义。
// @Summary 按单词查询释义
// @Description 根据英文拼写查询中文释义，可选择精确匹配或模糊匹配。
// @Tags dictionary
// @Produce json
// @Param spelling query string true "拼写"
// @Param exact query bool false "是否精确匹配"
// @Param limit query int false "返回数量"
// @Success 200 {object} response.Body{data=apiv1.DictionaryLookupResult}
// @Router /api/v1/dictionary/definitions [get]
func (h *DictionaryHandler) Lookup(c *gin.Context) {
	var q dtoV1.DictionaryLookupQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}

	result, err := h.service.LookupMeanings(dto.DictionaryLookupRequest{
		Spelling: q.Spelling,
		Exact:    q.Exact,
		Limit:    q.Limit,
	})
	if err != nil {
		handleDictionaryError(c, err)
		return
	}
	response.Success(c, toLookupResult(result))
}

// ListWords 分页查询单词。
// @Summary 分页查询单词
// @Description 按关键词、拼写、释义、词书和词频范围分页查询单词。
// @Tags dictionary
// @Produce json
// @Param keyword query string false "关键词"
// @Param spelling query string false "拼写"
// @Param paraphrase query string false "释义"
// @Param bookId query int false "词书 ID"
// @Param minFrequency query number false "最小词频"
// @Param maxFrequency query number false "最大词频"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Body{data=apiv1.DictionaryWordPage}
// @Router /api/v1/dictionary/words [get]
func (h *DictionaryHandler) ListWords(c *gin.Context) {
	var q dtoV1.DictionaryWordListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}

	result, err := h.service.ListWords(dto.DictionaryWordListRequest{
		Keyword:      q.Keyword,
		Spelling:     q.Spelling,
		Paraphrase:   q.Paraphrase,
		BookID:       q.BookID,
		MinFrequency: q.MinFrequency,
		MaxFrequency: q.MaxFrequency,
		PageQuery: dto.PageQuery{
			Page:     q.Page,
			PageSize: q.PageSize,
		},
	})
	if err != nil {
		handleDictionaryError(c, err)
		return
	}

	items := make([]dtoV1.DictionaryWordListItem, len(result.List))
	for i, item := range result.List {
		items[i] = toWordListItem(item)
	}
	response.Success(c, newPage(items, result.Total, q.PageQuery))
}

// GetWord 获取单词详情。
// @Summary 获取单词详情
// @Description 根据单词 ID 获取单词基础信息、所属词书和例句。
// @Tags dictionary
// @Produce json
// @Param wordId path int true "单词 ID"
// @Success 200 {object} response.Body{data=apiv1.DictionaryWordDetail}
// @Router /api/v1/dictionary/words/{wordId} [get]
func (h *DictionaryHandler) GetWord(c *gin.Context) {
	wordID, ok := pathUint(c, "wordId")
	if !ok {
		return
	}

	result, err := h.service.GetWordDetail(wordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, apperror.New(apperror.CodeNotFound, "word not found"))
			return
		}
		handleDictionaryError(c, err)
		return
	}
	response.Success(c, toWordDetail(result))
}

// GetWordBySpelling 按拼写获取单词详情。
// @Summary 按拼写获取单词详情
// @Description 根据英文拼写精确获取单词基础信息、所属词书和例句。
// @Tags dictionary
// @Produce json
// @Param spelling path string true "拼写"
// @Success 200 {object} response.Body{data=apiv1.DictionaryWordDetail}
// @Router /api/v1/dictionary/words/by-spelling/{spelling} [get]
func (h *DictionaryHandler) GetWordBySpelling(c *gin.Context) {
	result, err := h.service.GetWordBySpelling(c.Param("spelling"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, apperror.New(apperror.CodeNotFound, "word not found"))
			return
		}
		handleDictionaryError(c, err)
		return
	}
	response.Success(c, toWordDetail(result))
}

// ListBooks 分页查询词书。
// @Summary 分页查询词书
// @Description 按词书名称和状态分页查询词书。
// @Tags dictionary
// @Produce json
// @Param keyword query string false "关键词"
// @Param status query int false "状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Body{data=apiv1.DictionaryBookPage}
// @Router /api/v1/dictionary/books [get]
func (h *DictionaryHandler) ListBooks(c *gin.Context) {
	var q dtoV1.DictionaryBookListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}

	result, err := h.service.ListBooks(dto.DictionaryBookListRequest{
		Keyword: q.Keyword,
		Status:  q.Status,
		PageQuery: dto.PageQuery{
			Page:     q.Page,
			PageSize: q.PageSize,
		},
	})
	if err != nil {
		handleDictionaryError(c, err)
		return
	}

	items := make([]dtoV1.DictionaryBookListItem, len(result.List))
	for i, item := range result.List {
		items[i] = toBookListItem(item)
	}
	response.Success(c, newPage(items, result.Total, q.PageQuery))
}

// ListBookWords 分页查询词书单词。
// @Summary 分页查询词书单词
// @Description 根据词书 ID 分页查询该词书下的单词，支持关键词搜索。
// @Tags dictionary
// @Produce json
// @Param bookId path int true "词书 ID"
// @Param keyword query string false "关键词"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Body{data=apiv1.DictionaryWordPage}
// @Router /api/v1/dictionary/books/{bookId}/words [get]
func (h *DictionaryHandler) ListBookWords(c *gin.Context) {
	bookID, ok := pathUint(c, "bookId")
	if !ok {
		return
	}

	var q dtoV1.DictionaryBookWordsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}

	result, err := h.service.ListBookWords(bookID, dto.DictionaryBookWordsRequest{
		Keyword: q.Keyword,
		PageQuery: dto.PageQuery{
			Page:     q.Page,
			PageSize: q.PageSize,
		},
	})
	if err != nil {
		handleDictionaryError(c, err)
		return
	}

	items := make([]dtoV1.DictionaryWordListItem, len(result.List))
	for i, item := range result.List {
		items[i] = toWordListItem(item)
	}
	response.Success(c, newPage(items, result.Total, q.PageQuery))
}

func handleDictionaryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidParam):
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
	case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
		response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
	default:
		response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
	}
}

func toWordListItem(item dto.DictionaryWordListItem) dtoV1.DictionaryWordListItem {
	return dtoV1.DictionaryWordListItem{
		WordID:     item.WordID,
		Spelling:   item.Spelling,
		UKPhonetic: item.UKPhonetic,
		USPhonetic: item.USPhonetic,
		Paraphrase: item.Paraphrase,
		Frequency:  item.Frequency,
	}
}

func toBookListItem(item dto.DictionaryBookListItem) dtoV1.DictionaryBookListItem {
	return dtoV1.DictionaryBookListItem{
		BookID:   item.BookID,
		BookName: item.BookName,
		VocCount: item.VocCount,
		Status:   item.Status,
	}
}

func toWordDetail(detail dto.DictionaryWordDetail) dtoV1.DictionaryWordDetail {
	books := make([]dtoV1.DictionaryBookSimple, len(detail.Books))
	for i, book := range detail.Books {
		books[i] = toBookSimple(book)
	}
	examples := make([]dtoV1.DictionaryWordExampleItem, len(detail.Examples))
	for i, example := range detail.Examples {
		examples[i] = toWordExample(example)
	}
	return dtoV1.DictionaryWordDetail{
		Word:     toWordListItem(detail.Word),
		Books:    books,
		Examples: examples,
	}
}

func toBookSimple(book dto.DictionaryBookSimple) dtoV1.DictionaryBookSimple {
	return dtoV1.DictionaryBookSimple{
		BookID:   book.BookID,
		BookName: book.BookName,
	}
}

func toWordExample(example dto.DictionaryWordExampleItem) dtoV1.DictionaryWordExampleItem {
	return dtoV1.DictionaryWordExampleItem{
		ExaPID:  example.ExaPID,
		EN:      example.EN,
		CN:      example.CN,
		Heat:    example.Heat,
		AddDate: example.AddDate,
	}
}

func toLookupResult(result dto.DictionaryLookupResult) dtoV1.DictionaryLookupResult {
	items := make([]dtoV1.DictionaryWordListItem, len(result.Items))
	for i, item := range result.Items {
		items[i] = toWordListItem(item)
	}
	return dtoV1.DictionaryLookupResult{
		Spelling: result.Spelling,
		Exact:    result.Exact,
		Total:    result.Total,
		Items:    items,
	}
}
