package models

type Tabler interface {
	TableName() string
}

type User struct {
	ID               int
	Username         string `json:"username"`
	Email            string `json:"email"`
	Fullname         string `json:"fullname"`
	ProfileBio       string `json:"profileBio"`
	Password         string `json:"password"`
	Salt             string `json:"salt"`
	Token            string `json:"token"`
	IsSubscribed     bool   `json:"isSubscribed" gorm:"default:false"`
	VerificationCode string `json:"verificationCode" gorm:"-"`
}

type Word struct {
	ID          int
	Word        string
	Phonetics   string
	Definitions []Definition
}

type Definition struct {
	ID           int
	PartOfSpeech string
	Definition   string
	Example      string
	WordID       int
}

type VerificationCode struct {
	ID        int
	Code      string
	Email     string
	Attempts  int
	StartTime int64
}

type CollectionWord struct {
	ID           int
	CollectionID int
	WordID       int
	Priority     int
}

type Collection struct {
	ID     int
	Name   string
	Words  []CollectionWord
	UserID int
}
