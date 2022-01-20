package models

import "gorm.io/gorm"

type Tabler interface {
	TableName() string
}

type User struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
	Token    string `json:"token"`
}

func (User) TableName() string {
	return "users"
}
