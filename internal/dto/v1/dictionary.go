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

type DictionaryWordListItem struct {
	WordID     uint    `json:"wordId"`
	Spelling   string  `json:"spelling"`
	UKPhonetic string  `json:"ukPhonetic"`
	USPhonetic string  `json:"usPhonetic"`
	Paraphrase string  `json:"paraphrase"`
	Frequency  float64 `json:"frequency"`
}

type DictionaryWordExampleItem struct {
	ExaPID  int    `json:"exaPid"`
	EN      string `json:"en"`
	CN      string `json:"cn"`
	Heat    *int   `json:"heat"`
	AddDate string `json:"addDate"`
}

type DictionaryBookSimple struct {
	BookID   uint   `json:"bookId"`
	BookName string `json:"bookName"`
}

type DictionaryWordDetail struct {
	Word     DictionaryWordListItem      `json:"word"`
	Books    []DictionaryBookSimple      `json:"books"`
	Examples []DictionaryWordExampleItem `json:"examples"`
}

type DictionaryBookListItem struct {
	BookID   uint   `json:"bookId"`
	BookName string `json:"bookName"`
	VocCount *int   `json:"vocCount"`
	Status   *int   `json:"status"`
}

type DictionaryLookupResult struct {
	Spelling string                   `json:"spelling"`
	Exact    bool                     `json:"exact"`
	Total    int                      `json:"total"`
	Items    []DictionaryWordListItem `json:"items"`
}
