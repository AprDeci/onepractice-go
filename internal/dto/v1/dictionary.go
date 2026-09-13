package apiv1

type DictionaryWordListQuery struct {
	Keyword      string   `form:"keyword" json:"keyword"`
	Spelling     string   `form:"spelling" json:"spelling"`
	Paraphrase   string   `form:"paraphrase" json:"paraphrase"`
	BookID       *uint    `form:"bookId" json:"bookId"`
	MinFrequency *float64 `form:"minFrequency" json:"minFrequency"`
	MaxFrequency *float64 `form:"maxFrequency" json:"maxFrequency"`
	PageQuery
}

type DictionaryBookListQuery struct {
	Keyword string `form:"keyword" json:"keyword"`
	Status  *int   `form:"status" json:"status"`
	PageQuery
}

type DictionaryBookWordsQuery struct {
	Keyword string `form:"keyword" json:"keyword"`
	PageQuery
}

type DictionaryLookupQuery struct {
	Spelling string `form:"spelling" json:"spelling" binding:"required"`
	Exact    bool   `form:"exact" json:"exact"`
	Limit    int    `form:"limit" json:"limit"`
}
