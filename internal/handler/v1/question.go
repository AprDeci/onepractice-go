package v1

import (
	"errors"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	"onepractice-golang/internal/dto"
	dtoV1 "onepractice-golang/internal/dto/v1"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

type QuestionHandler struct {
	service *service.QuestionService
}

func NewQuestionHandler(svc *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{service: svc}
}

// List 按试卷获取题目，支持题型过滤与按 Part 分组。
// @Summary 按试卷获取题目
// @Description 返回指定试卷下的全部题目；默认 data 为题目数组，传 groupBy=part 时 data 为 ExamQuestion 分组结构。
// @Tags question
// @Produce json
// @Param paperId path int true "试卷 ID"
// @Param type query string false "题型"
// @Param groupBy query string false "传 part 按 Part 分组"
// @Success 200 {object} response.Body{data=[]model.Question}
// @Router /api/v1/papers/{paperId}/questions [get]
func (h *QuestionHandler) List(c *gin.Context) {
	paperID, ok := pathInt(c, "paperId")
	if !ok {
		return
	}

	questionType := c.Query("type")
	if c.Query("groupBy") == "part" {
		questions, err := h.service.SplitByPart(paperID)
		if err != nil {
			writeQuestionError(c, err)
			return
		}
		response.Success(c, toExamQuestion(questions))
		return
	}

	if questionType != "" {
		questions, err := h.service.ByPaperIDAndType(paperID, questionType)
		if err != nil {
			writeQuestionError(c, err)
			return
		}
		response.Success(c, questions)
		return
	}

	questions, err := h.service.ByPaperID(paperID)
	if err != nil {
		writeQuestionError(c, err)
		return
	}
	response.Success(c, questions)
}

// Answers 获取试卷答案。
// @Summary 获取试卷答案
// @Description 返回指定试卷的全部答案，并按答案序号排序。
// @Tags question
// @Produce json
// @Param paperId path int true "试卷 ID"
// @Success 200 {object} response.Body{data=apiv1.AnswersResponse}
// @Router /api/v1/papers/{paperId}/answers [get]
func (h *QuestionHandler) Answers(c *gin.Context) {
	paperID, ok := pathInt(c, "paperId")
	if !ok {
		return
	}

	answers, err := h.service.Answers(paperID)
	if err != nil {
		writeQuestionError(c, err)
		return
	}
	response.Success(c, toAnswersResponse(answers))
}

// Practice 随机获取指定数量的专项训练题。
// @Summary 获取专项训练题目
// @Description 无需登录，返回随机试卷中的指定题型、题组和答案。
// @Tags question
// @Accept json
// @Produce json
// @Param request body apiv1.PracticeQuestionRequest true "专项训练参数"
// @Success 200 {object} response.Body{data=apiv1.PracticeQuestionResponse}
// @Router /api/v1/questions/practice [post]
func (h *QuestionHandler) Practice(c *gin.Context) {
	var req dtoV1.PracticeQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}

	questions, err := h.service.Practice(req.QuestionType, req.UnitCount)
	if err != nil {
		writeQuestionError(c, err)
		return
	}
	response.Success(c, toPracticeResponse(questions))
}

func writeQuestionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidParam), errors.Is(err, service.ErrInvalidQuestionType), errors.Is(err, service.ErrInvalidPracticeUnitCount):
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
	case errors.Is(err, service.ErrPracticeQuestionsNotFound):
		response.Error(c, apperror.New(apperror.CodeNotFound, "资源不存在"))
	case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
		response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
	default:
		response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
	}
}

func toQuestionPart(part dto.QuestionPart) dtoV1.QuestionPart {
	return dtoV1.QuestionPart{Questions: part.Questions}
}

func toExamQuestion(src dto.ExamQuestion) dtoV1.ExamQuestion {
	parts := make([]dtoV1.QuestionPart, 0, len(src.QuestionParts))
	for _, part := range src.QuestionParts {
		parts = append(parts, toQuestionPart(part))
	}
	return dtoV1.ExamQuestion{PaperID: src.PaperID, QuestionParts: parts}
}

func toAnswersResponse(src dto.AnswersResponse) dtoV1.AnswersResponse {
	return dtoV1.AnswersResponse{PaperID: src.PaperID, Answers: src.Answers}
}

func toPracticeGroup(src dto.PracticeQuestionGroup) dtoV1.PracticeQuestionGroup {
	parts := make([]dtoV1.QuestionPart, 0, len(src.QuestionParts))
	for _, part := range src.QuestionParts {
		parts = append(parts, toQuestionPart(part))
	}
	return dtoV1.PracticeQuestionGroup{PaperID: src.PaperID, QuestionParts: parts, Answers: src.Answers}
}

func toPracticeResponse(src dto.PracticeQuestionResponse) dtoV1.PracticeQuestionResponse {
	groups := make([]dtoV1.PracticeQuestionGroup, 0, len(src.Groups))
	for _, group := range src.Groups {
		groups = append(groups, toPracticeGroup(group))
	}
	return dtoV1.PracticeQuestionResponse{QuestionType: src.QuestionType, Groups: groups}
}
