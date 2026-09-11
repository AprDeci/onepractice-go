package dto

type PaperQueryRequest struct {
	Page int    `json:"page" binding:"required,min=1"`
	Size int    `json:"size" binding:"required,min=1,max=100"`
	Type string `json:"type"`
	Year int    `json:"year"`
}
