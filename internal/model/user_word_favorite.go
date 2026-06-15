package model

import "time"

type UserWordFavorite struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"column:user_id" json:"user_id"`
	WordID    uint      `gorm:"column:wordid" json:"wordid"`
	PaperID   *int      `gorm:"column:paper_id" json:"paper_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserWordFavorite) TableName() string { return "user_word_favorites" }
