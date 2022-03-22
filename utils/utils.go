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
	"os"
	"sort"
	"strings"

	gomail "gopkg.in/gomail.v2"
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

func MailVerificationCode(code string, email string) error {
	m := gomail.NewMessage()
	fromEmail := os.Getenv("EMAIL")
	password := os.Getenv("EMAIL_PASSWORD")

	m.SetHeader("From", fromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Verification code")
	m.SetBody("text/plain", "Your verification code for Mercury: "+code)

	d := gomail.NewDialer("smtp.yandex.ru", 587, fromEmail, password)

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func ParseUser(user *models.User, req *http.Request) error {
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(user)
	return err
}

func AuthenticateToken(w *http.ResponseWriter, req *http.Request, user *models.User, token string) bool {
	result := database.DB.Table("users").Where("token = ?", token).First(user)
	if result.RowsAffected == 0 || result.Error != nil {
		(*w).WriteHeader(http.StatusBadRequest)
		fmt.Fprint(*w, `{"error": "Invalid token"}`)
		return false
	}
	return true
}

func GenerateWordsJSON(words []models.CollectionWord, user models.User) []map[string]any {
	var wordsToReturn []map[string]any
	for _, word := range words {
		definitions := make(map[string][]map[string]string)

		for _, element := range word.Word.Definitions {
			definitions[element.PartOfSpeech] = append(definitions[element.PartOfSpeech], map[string]string{
				"definition": element.Definition,
				"example":    element.Example,
			})
		}
		var priority models.Priority
		database.DB.Model(&models.Priority{}).Where("collection_word_id = ? and user_id = ?", word.ID, user.ID).Select("priority").Find(&priority)
		result := make(map[string]any)
		result["word"] = word.Word.Word
		result["definitions"] = definitions
		result["phonetics"] = word.Word.Phonetics
		result["priority"] = priority.Priority
		wordsToReturn = append(wordsToReturn, result)
	}
	sort.Slice(wordsToReturn, func(i, j int) bool {
		return wordsToReturn[i]["priority"].(int) < wordsToReturn[j]["priority"].(int)
	})
	return wordsToReturn
}
