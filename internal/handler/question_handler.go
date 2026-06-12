package handler

import (
	"net/http"
	"strconv"

	"onepractice-golang/internal/response"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

type QuestionHandler struct {
	service *service.QuestionService
}

func NewQuestionHandler(service *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{service: service}
}

// ByPaperID 按试卷 ID 获取题目。
// @Summary 按试卷 ID 获取题目
// @Description 返回指定试卷下的全部题目。
// @Tags question
// @Produce json
// @Param id query int true "试卷 ID"
// @Success 200 {object} response.Body
// @Router /api/question/getById [get]
func (h *QuestionHandler) ByPaperID(c *gin.Context) {
	id, ok := queryInt(c, "id")
	if !ok {
		return
	}

	questions, err := h.service.ByPaperID(id)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, questions)
}

// ByPaperIDAndType 按试卷 ID 和题型获取题目。
// @Summary 按试卷 ID 和题型获取题目
// @Description 返回指定试卷下指定题型的题目。
// @Tags question
// @Produce json
// @Param id query int true "试卷 ID"
// @Param type query string true "题型"
// @Success 200 {object} response.Body
// @Router /api/question/getByType [get]
func (h *QuestionHandler) ByPaperIDAndType(c *gin.Context) {
	id, ok := queryInt(c, "id")
	if !ok {
		return
	}

	questions, err := h.service.ByPaperIDAndType(id, c.Query("type"))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, questions)
}

// SplitByPart 按 Part 分组获取题目。
// @Summary 按 Part 分组获取题目
// @Description 返回指定试卷题目，并按 partName 分组。Part II 内按 questionOrder 排序。
// @Tags question
// @Produce json
// @Param id query int true "试卷 ID"
// @Success 200 {object} response.Body
// @Router /api/question/getAllByIdSplitByPart [get]
func (h *QuestionHandler) SplitByPart(c *gin.Context) {
	id, ok := queryInt(c, "id")
	if !ok {
		return
	}

	questions, err := h.service.SplitByPart(id)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, questions)
}

// Answers 获取试卷答案。
// @Summary 获取试卷答案
// @Description 返回指定试卷的全部答案，并按答案序号排序。
// @Tags question
// @Produce json
// @Param id query int true "试卷 ID"
// @Success 200 {object} response.Body
// @Router /api/question/getAnswersByPaperId [get]
func (h *QuestionHandler) Answers(c *gin.Context) {
	id, ok := queryInt(c, "id")
	if !ok {
		return
	}

	answers, err := h.service.Answers(id)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, answers)
}

func queryInt(c *gin.Context, key string) (int, bool) {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		response.Error(c, http.StatusBadRequest, key+" must be integer")
		return 0, false
	}
	return value, true
}
