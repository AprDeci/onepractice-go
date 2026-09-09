package apiv1

import "time"

type FavoriteWordRequest struct {
	WordID  uint `json:"wordId" form:"wordId" binding:"required"`
	PaperID *int `json:"paperId" form:"paperId"`
}

type FavoriteWordListQuery struct {
	Keyword string `form:"keyword" json:"keyword"`
	PageQuery
}

type FavoriteStatusResponse struct {
	WordID    uint `json:"wordId"`
	Favorited bool `json:"favorited"`
}

type CollectedWordItem struct {
	ID         uint      `json:"id"`
	FavoriteID uint64    `json:"favoriteId"`
	WordID     uint      `json:"wordId"`
	Word       string    `json:"word"`
	Spelling   string    `json:"spelling"`
	UKPhonetic string    `json:"ukPhonetic"`
	USPhonetic string    `json:"usPhonetic"`
	Paraphrase string    `json:"paraphrase"`
	Frequency  float64   `json:"frequency"`
	PaperID    *int      `json:"paperId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type CollectedWordList struct {
	Total int64               `json:"total"`
	Data  []CollectedWordItem `json:"data"`
}
