package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mercury/models"
	"mercury/utils"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func CheckToken(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	var user models.User
	err := utils.ParseAndAuthenticate(&user, &w, req)
	if err != nil {
		return
	}
}

func GetUserData(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	vars := mux.Vars(req)
	user := models.User{Token: vars["token"]}

	if err := utils.AuthenticateToken(&user); err != nil {
		// Invalid token
		w.WriteHeader(http.StatusBadRequest)
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
	utils.EnableCors(&w)

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
