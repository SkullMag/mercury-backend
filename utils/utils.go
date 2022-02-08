package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"mercury/database"
	"mercury/models"
	"net/http"
	"strings"
)

// https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomStringURLSafe returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomStringURLSafe(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	return hex.EncodeToString(b), err
}

func GenerateVerificationCode() (string, error) {
	code := make([]string, 4)
	for i := 0; i < 4; i++ {
		randInt, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = randInt.String()
	}
	return strings.Join(code, ""), nil
}

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func ParseUser(user *models.User, req *http.Request) error {
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(user)
	return err
}

func AuthenticateToken(user *models.User) error {
	result := database.DB.Table("users").Where("token = ?", user.Token).First(user)
	return result.Error
}

func ParseAndAuthenticate(user *models.User, w *http.ResponseWriter, req *http.Request) error {
	err := ParseUser(user, req)
	if err != nil {
		(*w).WriteHeader(http.StatusBadRequest)
		fmt.Fprint(*w, "{\"error\": \"error occurred while parsing input data\"}")
		return err
	}

	isAuth := AuthenticateToken(user)
	if isAuth != nil {
		(*w).WriteHeader(http.StatusBadRequest)
		fmt.Fprint(*w, "{\"error\": \"invalid token\"}")
		return isAuth
	}
	return nil
}
