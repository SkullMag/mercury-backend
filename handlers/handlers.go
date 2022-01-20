package handlers

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mercury/database"
	"mercury/models"
	"mercury/utils"
	"net/http"
)

func Register(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)
	errorResponse := map[string]string{"error": ""}
	if req.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	var user models.User
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&user)
	if err != nil || user.Username == "" || user.Password == "" {
		w.WriteHeader(400)
		errorResponse["error"] = "Username or password was not provided"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprintf(w, string(response))
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
		w.WriteHeader(400)
		errorResponse["error"] = "Username is not unique"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprintf(w, string(response))
		return
	}
}

func Login(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)
	if req.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	var user models.User
	var dbUser models.User
	errorResponse := map[string]string{"error": ""}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&user)
	if err != nil {
		w.WriteHeader(400)
	}
	result := database.DB.Where("username = ?", user.Username).First(&dbUser)
	if result.Error != nil {
		w.WriteHeader(400)
		errorResponse["error"] = "Username was not found"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprintf(w, string(response))
		return
	}
	hasher := sha512.New()
	hasher.Write([]byte(user.Password + dbUser.Salt))
	if hex.EncodeToString(hasher.Sum(nil)) == dbUser.Password {
		response, _ := json.Marshal(map[string]string{"token": dbUser.Token})
		fmt.Fprintf(w, string(response))
	} else {
		w.WriteHeader(400)
		errorResponse["error"] = "Wrong password"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprintf(w, string(response))
	}
}

func CheckToken(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	var user models.User
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&user)
	if user.Token == "" {
		w.WriteHeader(400)
		return
	}
	result := database.DB.Where("token = ?", user.Token).First(&user)
	if result.Error != nil {
		w.WriteHeader(400)
	}
}
