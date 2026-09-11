package dto

type DictionaryWordListRequest struct {
	Keyword      string   `form:"keyword" json:"keyword"`
	Spelling     string   `form:"spelling" json:"spelling"`
	Paraphrase   string   `form:"paraphrase" json:"paraphrase"`
	BookID       *uint    `form:"bookid" json:"bookid"`
	MinFrequency *float64 `form:"min_frequency" json:"min_frequency"`
	MaxFrequency *float64 `form:"max_frequency" json:"max_frequency"`
	PageQuery
}

func (r *DictionaryWordListRequest) Normalize() {
	r.PageQuery.Normalize()
}

type DictionaryBookListRequest struct {
	Keyword string `form:"keyword" json:"keyword"`
	Status  *int   `form:"status" json:"status"`
	PageQuery
}

func (r *DictionaryBookListRequest) Normalize() {
	r.PageQuery.Normalize()
}

type DictionaryBookWordsRequest struct {
	Keyword string `form:"keyword" json:"keyword"`
	PageQuery
}

func (r *DictionaryBookWordsRequest) Normalize() {
	r.PageQuery.Normalize()
}

type DictionaryLookupRequest struct {
	Spelling string `form:"spelling" binding:"required" json:"spelling"`
	Exact    bool   `form:"exact" json:"exact"`
	Limit    int    `form:"limit" json:"limit"`
}

func (r *DictionaryLookupRequest) Normalize() {
	if r.Limit <= 0 {
		r.Limit = 20
	}
	if r.Limit > 100 {
		r.Limit = 100
	}
}
