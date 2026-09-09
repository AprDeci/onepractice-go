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

type PaperHandler struct{ service *service.PaperService }

func NewPaperHandler(svc *service.PaperService) *PaperHandler { return &PaperHandler{service: svc} }

// List 获取试卷列表。
// @Summary 获取试卷列表
// @Description 分页查询试卷；include=rating 时返回含评分与题量的数据。
// @Tags paper
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param type query string false "试卷类型"
// @Param year query int false "年份"
// @Param include query string false "传 rating 返回评分信息"
// @Success 200 {object} response.Body
// @Router /api/v1/papers [get]
func (h *PaperHandler) List(c *gin.Context) {
	var q dtoV1.PaperQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	q.PageQuery = q.PageQuery.Normalize()
	query := dto.PaperQueryRequest{Page: q.Page, Size: q.PageSize, Type: q.Type, Year: q.Year}

	if q.Include == "rating" {
		result, err := h.service.PageWithRating(query)
		if err != nil {
			paperError(c, err)
			return
		}
		items := make([]dtoV1.PaperWithRating, 0, len(result.Data))
		for _, paper := range result.Data {
			items = append(items, toPaperWithRating(paper))
		}
		response.Success(c, newPage(items, result.Total, q.PageQuery))
		return
	}

	result, err := h.service.Page(query)
	if err != nil {
		paperError(c, err)
		return
	}
	response.Success(c, newPage(result.Data, result.Total, q.PageQuery))
}

// Intro 获取试卷简介。
// @Summary 获取试卷简介
// @Description 返回指定试卷的题型、时长、难度与各部分题量。
// @Tags paper
// @Produce json
// @Param paperId path int true "试卷 ID"
// @Success 200 {object} response.Body
// @Router /api/v1/papers/{paperId}/intro [get]
func (h *PaperHandler) Intro(c *gin.Context) {
	paperID, ok := pathInt(c, "paperId")
	if !ok {
		return
	}
	intro, err := h.service.Intro(paperID)
	if err != nil {
		paperError(c, err)
		return
	}
	response.Success(c, toPaperIntro(intro))
}

// Types 获取试卷类型列表。
// @Summary 获取试卷类型列表
// @Description 返回全部可用的试卷类型。
// @Tags paper
// @Produce json
// @Success 200 {object} response.Body
// @Router /api/v1/paper-types [get]
func (h *PaperHandler) Types(c *gin.Context) {
	types, err := h.service.Types()
	if err != nil {
		paperError(c, err)
		return
	}
	response.Success(c, types)
}

func toPaperWithRating(paper dto.PaperWithRating) dtoV1.PaperWithRating {
	return dtoV1.PaperWithRating{
		PaperID:       paper.PaperID,
		PaperName:     paper.PaperName,
		ExamYear:      paper.ExamYear,
		ExamMonth:     paper.ExamMonth,
		Version:       paper.Version,
		TotalTime:     paper.TotalTime,
		Type:          paper.Type,
		QuestionCount: paper.QuestionCount,
		Rating:        paper.Rating,
		Number:        paper.Number,
	}
}

func toPaperIntro(intro dto.PaperIntro) dtoV1.PaperIntro {
	return dtoV1.PaperIntro{
		PaperName:            intro.PaperName,
		ExamYear:             intro.ExamYear,
		ExamMonth:            intro.ExamMonth,
		PaperType:            intro.PaperType,
		PaperTime:            intro.PaperTime,
		Difficulty:           intro.Difficulty,
		SectionCount:         intro.SectionCount,
		SectionQuestionCount: intro.SectionQuestionCount,
	}
}

func paperError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidParam):
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
	case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
		response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
	default:
		response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
	}
}
