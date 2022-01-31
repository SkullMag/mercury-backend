package handlers

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mercury/database"
	"mercury/models"
	"mercury/utils"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func Register(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	errorResponse := map[string]string{"error": ""}
	var user models.User

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&user)
	if err != nil || user.Username == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse["error"] = "Username or password was not provided"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprint(w, string(response))
		return
	}
	token, _ := utils.GenerateRandomStringURLSafe(32)
	salt, _ := utils.GenerateRandomStringURLSafe(32)
	hasher := sha512.New()
	hasher.Write([]byte(user.Password + salt))
	user.Password = hex.EncodeToString(hasher.Sum(nil))
	user.Token = token
	user.Salt = salt
	result := database.DB.Create(&user)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse["error"] = "Username is not unique"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprint(w, string(response))
		return
	}
}

func Login(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	var user models.User
	var dbUser models.User
	errorResponse := map[string]string{"error": ""}
	err := utils.ParseUser(&user, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	result := database.DB.Where("username = ?", user.Username).First(&dbUser)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse["error"] = "Username was not found"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprint(w, string(response))
		return
	}
	hasher := sha512.New()
	hasher.Write([]byte(user.Password + dbUser.Salt))
	if hex.EncodeToString(hasher.Sum(nil)) == dbUser.Password {
		response, _ := json.Marshal(map[string]string{"token": dbUser.Token})
		fmt.Fprint(w, string(response))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse["error"] = "Wrong password"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprint(w, string(response))
	}
}

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

	// Token not provided
	if _, ok := vars["token"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := models.User{Token: vars["token"]}

	// Invalid token
	if err := utils.AuthenticateToken(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, _ := json.Marshal(map[string]string{
		"username": user.Username,
	})
	fmt.Fprint(w, string(response))
}

func GetUserProfilePicture(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	vars := mux.Vars(req)

	if _, ok := vars["username"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	path, _ := os.Getwd()
	fileBytes, err := ioutil.ReadFile(path + "/assets/" + vars["username"] + ".png")
	if err != nil {
		w.WriteHeader(404)
		fmt.Println(err.Error())
		return
	}
	w.Write(fileBytes)

}
