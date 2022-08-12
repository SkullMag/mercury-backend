package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mercury/database"
	"mercury/models"
	"mercury/utils"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func GetUserData(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var user models.User

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	response, _ := json.Marshal(map[string]string{
		"username":     user.Username,
		"fullname":     user.Fullname,
		"profileBio":   user.ProfileBio,
		"isSubscribed": strconv.FormatBool(user.IsSubscribed),
	})
	fmt.Fprint(w, string(response))
}

func GetUserProfilePicture(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	path, _ := os.Getwd()
	fileBytes, err := ioutil.ReadFile(path + "/assets/" + vars["username"] + ".png")
	if err != nil {
		w.WriteHeader(404)
		fmt.Println(err.Error())
		return
	}
	w.Write(fileBytes)

}

func GetUserDataByUsername(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var user models.User
	var requestedUser models.User

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	database.DB.Where("username = ?", vars["username"]).Find(&requestedUser)

	response, _ := json.Marshal(map[string]string{
		"username":     requestedUser.Username,
		"fullname":     requestedUser.Fullname,
		"profileBio":   requestedUser.ProfileBio,
		"isSubscribed": strconv.FormatBool(requestedUser.IsSubscribed),
	})
	fmt.Fprint(w, string(response))

}

func GetUserStats(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var user models.User
	var stats models.Stats

	var collectionID []int
	var collectionWords []models.CollectionWord
	var numberOfLearned int64 = 0

	database.DB.Where("username = ?", vars["username"]).Find(&user)
	database.DB.Table("collections").Select("id").Find(&collectionID)
	database.DB.Where("collection_id IN ?", collectionID).Find(&collectionWords)
	database.DB.Table("priorities").Where("user_id = ? and priority > 1", user.ID).Count(&numberOfLearned)
	// database.DB.Select("collection_words").Where("collection_id IN", collectionID).Count(&wordsCount)

	stats.CollectionsCount = len(collectionID)
	stats.WordsCount = len(collectionWords)
	stats.LearnedWordsCount = int(numberOfLearned)

	jsonEncoded, _ := json.Marshal(stats)
	fmt.Fprint(w, string(jsonEncoded))
	// database.DB.Select(&models.Collection{}).Where("authorUsername = ")

}

func DeleteProfile(w http.ResponseWriter, req *http.Request) {
    vars := mux.Vars(req)
    var user models.User
    var collectionIds []int

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}
    database.DB.Table("collections").Select("id").Where("user_id = ?", user.ID).Find(&collectionIds)

    database.DB.Exec("DELETE FROM collection_words WHERE collection_id IN ?", collectionIds)
    database.DB.Exec("DELETE FROM collections WHERE user_id = ?", user.ID)
    database.DB.Exec("DELETE FROM users WHERE id = ?", user.ID)

}
