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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func SignUp(w http.ResponseWriter, req *http.Request) {
	errorResponse := map[string]string{"error": ""}
	var user models.User
	var verificationCode models.VerificationCode

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&user)
	user.Username = strings.ToLower(user.Username)

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        errorResponse["error"] = "Error while decoding"
        response, _ := json.Marshal(errorResponse)
        fmt.Fprint(w, string(response))
        return
    }
	// Check body of request
	if user.Username == "" || user.Password == "" || user.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse["error"] = "Username or password was not provided " + user.Username + " " + user.Password + " " + user.Email
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

	if matched, _ := regexp.MatchString("^[A-z1-9]+$", user.Username); !matched {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Username should only contain ASCII characters"}`)
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
		fmt.Println(result.Error)
		w.WriteHeader(http.StatusBadRequest)
		errorResponse["error"] = "Username or email is not unique"
		response, _ := json.Marshal(errorResponse)
		fmt.Fprint(w, string(response))
		return
	}
}

func Login(w http.ResponseWriter, req *http.Request) {
	var user models.User
	var dbUser models.User
	errorResponse := map[string]string{"error": ""}
	err := utils.ParseUser(&user, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
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

func RequestVerificationCode(w http.ResponseWriter, req *http.Request) {
	var verificationCode models.VerificationCode
	var tempUser models.User

	vars := mux.Vars(req)

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
            fmt.Fprint(w, `{"error": "Try to resend token"}`)
			return
		}
		verificationCode.Code = code
		verificationCode.Email = vars["email"]
		verificationCode.StartTime = time.Now().Unix()
		verificationCode.Attempts = 0
		database.DB.Create(&verificationCode)
		utils.MailVerificationCode(verificationCode.Code, verificationCode.Email)
	} else {
		// diff := time.Since(time.Unix(verificationCode.StartTime, 0))
		// if diff.Seconds() < 60.0 {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	fmt.Fprint(w, `{"error": "Wait before requesting token"}`)
		// 	return
		// }
		code, err := utils.GenerateVerificationCode()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w, `{"error": "Try to resend token"}`)
			return
		}
		verificationCode.StartTime = time.Now().Unix()
		verificationCode.Attempts = 0
		verificationCode.Code = code
		mailError := utils.MailVerificationCode(verificationCode.Code, verificationCode.Email)
		if mailError != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error": "Try to resend token"}`)
		} else {
			database.DB.Save(&verificationCode)
		}
	}

}
