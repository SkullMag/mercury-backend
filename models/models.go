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

type Word struct {
	ID          int
	Word        string
	Phonetics   string
	Definitions []Definition
}

func (Word) TableName() string {
	return "words"
}

type Definition struct {
	ID           int
	PartOfSpeech string
	Definition   string
	Example      string
	WordID       int
}

func (Definition) TableName() string {
	return "definitions"
}
