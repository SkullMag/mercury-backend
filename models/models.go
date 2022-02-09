package models

type Tabler interface {
	TableName() string
}

type User struct {
	ID               int
	Username         string `json:"username"`
	Fullname         string `json:"fullname"`
	ProfileBio       string `json:"profileBio"`
	Password         string `json:"password"`
	Salt             string `json:"salt"`
	Token            string `json:"token"`
	IsSubscribed     bool   `json:"isSubscribed" gorm:"default:false"`
	VerificationCode string `gorm:"-"`
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

type VerificationCode struct {
	ID        int
	Code      string
	Email     string
	Attempts  int
	StartTime int64
}
