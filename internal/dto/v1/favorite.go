package apiv1

type FavoriteWordRequest struct {
	WordID  uint `json:"wordId" form:"wordId" binding:"required"`
	PaperID *int `json:"paperId" form:"paperId"`
}

type FavoriteWordListQuery struct {
	Keyword string `form:"keyword" json:"keyword"`
	PageQuery
}
