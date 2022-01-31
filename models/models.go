package models

type Tabler interface {
	TableName() string
}

type User struct {
	ID           int
	Username     string `json:"username"`
	Password     string `json:"password"`
	Salt         string `json:"salt"`
	Token        string `json:"token"`
	IsSubscribed bool   `json:"isSubscribed"`
}

func (User) TableName() string {
	return "users"
}
