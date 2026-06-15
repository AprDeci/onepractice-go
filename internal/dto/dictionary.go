package dto

type DictionaryPageRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

func (r *DictionaryPageRequest) Normalize() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 20
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
}

type DictionaryWordListRequest struct {
	Keyword      string   `form:"keyword" json:"keyword"`
	Spelling     string   `form:"spelling" json:"spelling"`
	Paraphrase   string   `form:"paraphrase" json:"paraphrase"`
	BookID       *uint    `form:"bookid" json:"bookid"`
	MinFrequency *float64 `form:"min_frequency" json:"min_frequency"`
	MaxFrequency *float64 `form:"max_frequency" json:"max_frequency"`
	DictionaryPageRequest
}

func (r *DictionaryWordListRequest) Normalize() {
	r.DictionaryPageRequest.Normalize()
}

type DictionaryBookListRequest struct {
	Keyword string `form:"keyword" json:"keyword"`
	Status  *int   `form:"status" json:"status"`
	DictionaryPageRequest
}

func (r *DictionaryBookListRequest) Normalize() {
	r.DictionaryPageRequest.Normalize()
}

type DictionaryBookWordsRequest struct {
	Keyword string `form:"keyword" json:"keyword"`
	DictionaryPageRequest
}

func (r *DictionaryBookWordsRequest) Normalize() {
	r.DictionaryPageRequest.Normalize()
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

type DictionaryWordListItem struct {
	WordID     uint    `gorm:"column:wordid" json:"wordid"`
	Spelling   string  `gorm:"column:spelling" json:"spelling"`
	UKPhonetic string  `gorm:"column:uk_phonetic" json:"uk_phonetic"`
	USPhonetic string  `gorm:"column:us_phonetic" json:"us_phonetic"`
	Paraphrase string  `gorm:"column:paraphrase" json:"paraphrase"`
	Frequency  float64 `gorm:"column:frequency" json:"frequency"`
}

type DictionaryWordExampleItem struct {
	ExaPID  int    `gorm:"column:exapid" json:"exapid"`
	EN      string `gorm:"column:en" json:"en"`
	CN      string `gorm:"column:cn" json:"cn"`
	Heat    *int   `gorm:"column:heat" json:"heat"`
	AddDate string `gorm:"column:adddate" json:"adddate"`
}

type DictionaryBookSimple struct {
	BookID   uint   `gorm:"column:bookid" json:"bookid"`
	BookName string `gorm:"column:bookname" json:"bookname"`
}

type DictionaryWordDetail struct {
	Word     DictionaryWordListItem      `json:"word"`
	Books    []DictionaryBookSimple      `json:"books"`
	Examples []DictionaryWordExampleItem `json:"examples"`
}

type DictionaryBookListItem struct {
	BookID   uint   `gorm:"column:bookid" json:"bookid"`
	BookName string `gorm:"column:bookname" json:"bookname"`
	VocCount *int   `gorm:"column:voccount" json:"voccount"`
	Status   *int   `gorm:"column:status" json:"status"`
}

type DictionaryPageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type DictionaryLookupResult struct {
	Spelling string                   `json:"spelling"`
	Exact    bool                     `json:"exact"`
	Total    int                      `json:"total"`
	Items    []DictionaryWordListItem `json:"items"`
}
