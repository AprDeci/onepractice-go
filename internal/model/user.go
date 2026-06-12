package model

import "time"

type User struct {
	ID        int64     `gorm:"column:id;primaryKey" json:"id"`
	Username  string    `gorm:"column:username" json:"username"`
	Password  string    `gorm:"column:password" json:"-"`
	Email     string    `gorm:"column:email" json:"email"`
	UserType  int       `gorm:"column:user_type" json:"userType"`
	CreatedAt time.Time `gorm:"column:create_time" json:"createTime"`
	UpdatedAt time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (User) TableName() string { return "user" }
