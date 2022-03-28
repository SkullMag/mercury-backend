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
