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

func (h *PaperHandler) All(c *gin.Context) {
	papers, err := h.service.All()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, papers)
}

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

func (h *PaperHandler) ByType(c *gin.Context) {
	papers, err := h.service.ByType(c.Query("type"))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, papers)
}

func (h *PaperHandler) Types(c *gin.Context) {
	types, err := h.service.Types()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, types)
}

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
