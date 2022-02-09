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
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func SignUp(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	errorResponse := map[string]string{"error": ""}
	var user models.User
	var verificationCode models.VerificationCode

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&user)

	// Check body of request
	if err != nil || user.Username == "" || user.Password == "" || user.Fullname == "" || user.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse["error"] = "Username or password was not provided"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprint(w, string(response))
		return
	}

	// Check verification code
	if user.VerificationCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "No verification code was provided"}`)
		return
	} else {
		database.DB.Where("email = ?", user.Email).First(&verificationCode)
		if verificationCode.Code == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "No verification code was requested"}`)
			return
		}
		if verificationCode.Attempts > 3 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "Request another verification code"}`)
			return
		}
		if verificationCode.Code != user.VerificationCode {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "Wrong verification code"}`)
			verificationCode.Attempts++
			database.DB.Save(&verificationCode)
			return
		}
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
		fmt.Println(result.Error)
		w.WriteHeader(http.StatusBadRequest)
		errorResponse["error"] = "Username or email is not unique"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprint(w, string(response))
		return
	}
}

func RequestVerificationCode(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	var verificationCode models.VerificationCode

	vars := mux.Vars(req)
	if _, ok := vars["email"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Email wasn't provided provided"}`)
		return
	}
	if _, ok := vars["username"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Username wasn't provided"}`)
		return
	}

	var tempUser models.User

	username := database.DB.Table("users").Where("username = ?", vars["username"]).Find(&tempUser)
	if username.RowsAffected > 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Username already exists"}`)
		return
	}

	email := database.DB.Table("users").Where("email = ?", vars["email"]).Find(&tempUser)
	if email.RowsAffected > 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Email is registered"}`)
		return
	}

	database.DB.Where("email = ?", vars["email"]).First(&verificationCode)
	if verificationCode.Code == "" {
		code, err := utils.GenerateVerificationCode()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		verificationCode.Code = code
		verificationCode.Email = vars["email"]
		verificationCode.StartTime = time.Now().Unix()
		verificationCode.Attempts = 0
		database.DB.Create(&verificationCode)
		utils.MailVerificationCode(verificationCode.Code, verificationCode.Email)
		return
	} else {
		diff := time.Since(time.Unix(verificationCode.StartTime, 0))
		if diff.Seconds() < 60.0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "Wait before requesting token"}`)
			return
		}
		code, err := utils.GenerateVerificationCode()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		verificationCode.StartTime = time.Now().Unix()
		verificationCode.Attempts = 0
		verificationCode.Code = code
		database.DB.Save(&verificationCode)
		mailError := utils.MailVerificationCode(verificationCode.Code, verificationCode.Email)
		if mailError != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error": "Try to resend token"}`)
		}
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
		response, _ := json.Marshal(map[string]string{
			"token":        dbUser.Token,
			"fullname":     dbUser.Fullname,
			"username":     dbUser.Username,
			"isSubscribed": strconv.FormatBool(dbUser.IsSubscribed),
			"profileBio":   dbUser.ProfileBio,
		})
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

	if _, ok := vars["token"]; !ok {
		// Token not provided
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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

func GetDefinition(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	vars := mux.Vars(req)

	if _, ok := vars["word"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var word models.Word
	database.DB.Where("word = ?", vars["word"]).First(&word)
	if word.Word == "" {
		// Word wasn't found in database
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error": "word was not found"}`)
		return
	}
	database.DB.Model(&word).Association("Definitions").Find(&word.Definitions)
	definitions := make(map[string][]map[string]string)

	for _, element := range word.Definitions {
		definitions[element.PartOfSpeech] = append(definitions[element.PartOfSpeech], map[string]string{
			"definition": element.Definition,
			"example":    element.Example,
		})
	}
	result := make(map[string]interface{})
	result["word"] = word.Word
	result["definitions"] = definitions
	result["phonetics"] = word.Phonetics

	jsonData, _ := json.Marshal(result)
	fmt.Fprint(w, string(jsonData))
}
