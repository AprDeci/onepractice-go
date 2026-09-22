package model

import "time"

type User struct {
	ID int64 `gorm:"column:id;primaryKey" json:"id"`
	// Nickname 是展示用昵称，不唯一、不参与登录；账号唯一标识是 Email。
	Nickname  string    `gorm:"column:nickname" json:"nickname"`
	Password  string    `gorm:"column:password" json:"-"`
	Email     string    `gorm:"column:email" json:"email"`
	UserType  int       `gorm:"column:user_type" json:"userType"`
	CreatedAt time.Time `gorm:"column:create_time" json:"createTime"`
	UpdatedAt time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (User) TableName() string { return "user" }
