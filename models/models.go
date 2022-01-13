package models


type User struct {
    ID int
    Username string `json:"username"`
    Password string `json:"password"`
    Salt string `json:"salt"`
    Token string `json:"token"`
}

func (User) TableName() string {
    return "users"
}