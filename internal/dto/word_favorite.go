package dto

type WordFavoriteRequest struct {
	WordID  uint   `form:"wordid" json:"wordid"`
	Word    string `form:"word" json:"word"`
	PaperID *int   `form:"paper_id" json:"paper_id"`
}

type WordFavoriteListRequest struct {
	Keyword string `form:"keyword" json:"keyword"`
	PageQuery
}
