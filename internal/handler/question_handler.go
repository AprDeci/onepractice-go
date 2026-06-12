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
