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
	IsCreated   bool         `json:"-"`
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
	ID               int            `json:"-"`
	UserID           int            `json:"-"`
	CollectionID     int            `json:"-"`
	CollectionWordID int            `json:"-"`
	Priority         int            `json:"priority"`
	CollectionWord   CollectionWord `gorm:"foreignKey:CollectionWordID"`
}

type CollectionWord struct {
	ID           int  `json:"-"`
	CollectionID int  `json:"-"`
	WordID       int  `json:"-"`
	Word         Word `json:"collectionWord"`
	Priority     int  `json:"priority" gorm:"-"`
}

type Collection struct {
	ID           int              `json:"id"`
	Name         string           `json:"name"`
	Words        []CollectionWord `json:"-"`
	UserID       int              `json:"-"`
	User         User             `json:"username"`
	Likes        int              `json:"likes"`
	WordCount    int              `json:"wordCount" gorm:"-"`
	ContainsWord bool             `json:"containsWord" gorm:"-"`
	IsFavourite  bool             `json:"isFavourite" gorm:"-"`
}

func (c Collection) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Likes        int    `json:"likes"`
		WordCount    int    `json:"wordCount"`
		Username     string `json:"username"`
		ContainsWord bool   `json:"containsWord"`
		IsFavourite  bool   `json:"isFavourite"`
	}{
		ID:           c.ID,
		Name:         c.Name,
		Likes:        c.Likes,
		WordCount:    len(c.Words),
		Username:     c.User.Username,
		ContainsWord: c.ContainsWord,
		IsFavourite:  c.IsFavourite,
	})
}

type Stats struct {
	CollectionsCount  int `json:"collectionsCount"`
	WordsCount        int `json:"wordsCount"`
	LearnedWordsCount int `json:"learnedWordsCount"`
}

type Favourite struct {
	ID           int
	UserID       int
	CollectionID int
	Collection   Collection
}

type WordToAdd struct {
	Term         string `json:"term"`
	Definition   string `json:"definition"`
	CollectionID int    `json:"collectionID"`
}
