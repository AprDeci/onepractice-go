package dto

import "time"

type WordFavoriteRequest struct {
	WordID  uint   `form:"wordid" json:"wordid"`
	Word    string `form:"word" json:"word"`
	PaperID *int   `form:"paper_id" json:"paper_id"`
}

type WordFavoriteListRequest struct {
	PageQuery
}

type CollectedWordItem struct {
	ID         uint      `gorm:"column:id" json:"id"`
	FavoriteID uint64    `gorm:"column:favorite_id" json:"favorite_id"`
	WordID     uint      `gorm:"column:wordid" json:"wordid"`
	Word       string    `gorm:"column:word" json:"word"`
	Spelling   string    `gorm:"column:spelling" json:"spelling"`
	UKPhonetic string    `gorm:"column:uk_phonetic" json:"uk_phonetic"`
	USPhonetic string    `gorm:"column:us_phonetic" json:"us_phonetic"`
	Paraphrase string    `gorm:"column:paraphrase" json:"paraphrase"`
	Frequency  float64   `gorm:"column:frequency" json:"frequency"`
	PaperID    *int      `gorm:"column:paper_id" json:"paper_id"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

type CollectedWordList struct {
	Total int64               `json:"total"`
	Data  []CollectedWordItem `json:"data"`
}
