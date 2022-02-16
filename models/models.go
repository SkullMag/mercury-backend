package models

import "encoding/json"

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
	Collections      []Collection
}

type Word struct {
	ID          int          `json:"-"`
	Word        string       `json:"word"`
	Phonetics   string       `json:"phonetics"`
	Definitions []Definition `json:"definitions"`
}

type Definition struct {
	ID           int    `json:"-"`
	PartOfSpeech string `json:"partOfSpeech"`
	Definition   string `json:"definition"`
	Example      string `json:"example"`
	WordID       int    `json:"-"`
}

type VerificationCode struct {
	ID        int
	Code      string
	Email     string
	Attempts  int
	StartTime int64
}

type Priority struct {
	ID           int `json:"-"`
	UserID       int `json:"-"`
	CollectionID int `json:"-"`
	WordID       int `json:"-"`
	Priority     int `json:"priority"`
}

type CollectionWord struct {
	ID           int  `json:"-"`
	CollectionID int  `json:"-"`
	WordID       int  `json:"-"`
	Word         Word `json:"collectionWord"`
}

type Collection struct {
	ID        int              `json:"-"`
	Name      string           `json:"name"`
	Words     []CollectionWord `json:"-"`
	UserID    int              `json:"-"`
	Likes     int              `json:"likes"`
	WordCount int              `json:"wordCount"`
}

func (c Collection) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name      string `json:"name"`
		Likes     int    `json:"likes"`
		WordCount int    `json:"wordCount"`
	}{
		Name:      c.Name,
		Likes:     c.Likes,
		WordCount: len(c.Words),
	})
}
