package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"mercury/database"
	"mercury/models"
	"mercury/utils"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Models for dictionaryapi.dev
type Definition struct {
	Definition string `json:"definition"`
	Example    string `json:"example"`
}

type Meaning struct {
	PartOfSpeech string       `json:"partOfSpeech"`
	Definitions  []Definition `json:"definitions"`
}

type Word struct {
	Word     string    `json:"word"`
	Meanings []Meaning `json:"meanings"`
}

type WordNotFoundError struct{}

func (err WordNotFoundError) Error() string {
	return "Word was not found"
}

// Function for getting words from dictionaryapi.dev
func getFreeDictionaryAPIDefinition(word string) error {
	var err error
	word = strings.ToLower(word)
	resp, err := http.Get(fmt.Sprintf("https://api.dictionaryapi.dev/api/v2/entries/en/%s", word))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var words []Word
	err = json.Unmarshal(body, &words)
	if err != nil {
		return err
	}

	if len(words) == 0 {
		return &WordNotFoundError{}
	}

	flag := false
	for _, w := range words {
		if w.Word == word {
			flag = true
			break
		}
	}

	if !flag {
		return &WordNotFoundError{}
	}

	var maxWordID int
	database.DB.Table("words").Select("max(id)").Find(&maxWordID)
	maxWordID++
	result := database.DB.Create(&models.Word{ID: maxWordID, Word: word})
	if result.Error != nil {
		return result.Error
	}

	var maxDefinitionID int
	database.DB.Table("definitions").Select("max(id)").Find(&maxDefinitionID)

	for _, w := range words {
		for _, meaning := range w.Meanings {
			for _, definition := range meaning.Definitions {
				maxDefinitionID++
				database.DB.Create(&models.Definition{
					ID:           maxDefinitionID,
					PartOfSpeech: meaning.PartOfSpeech,
					Definition:   definition.Definition,
					Example:      definition.Example,
					WordID:       maxWordID,
				})
			}
		}
	}
	return nil
}

func GetDefinition(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	var word models.Word
	database.DB.Where("word = ? AND is_created = 0", vars["word"]).First(&word)
	if word.Word == "" {
		if err := getFreeDictionaryAPIDefinition(vars["word"]); err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error": "Word was not found"}`)
			return
		}
		var newWord models.Word
		res := database.DB.Where("word = ?", strings.ToLower(vars["word"])).First(&newWord)
		if res.RowsAffected > 0 {
			word = newWord
		}
	}
	if word.Word == "" {
		// Word wasn't found in database
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error": "Word was not found"}`)
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
	result := make(map[string]any)
	result["word"] = word.Word
	result["definitions"] = definitions
	result["phonetics"] = word.Phonetics

	jsonData, _ := json.Marshal(result)
	fmt.Fprint(w, string(jsonData))
}

func AddWord(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var user models.User
	var wordToAdd models.WordToAdd
	var collection models.Collection

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&wordToAdd); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Invalid request body"}`)
		return
	}

	if res := database.DB.Where("id = ? AND user_id = ?", wordToAdd.CollectionID, user.ID).Find(&collection); res.RowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error": "Collection was not found"}`)
		return
	}

	if wordToAdd.Term == "" || wordToAdd.Definition == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Name or definition was not specified"}`)
		return
	}

	// TODO: replace with autoincrement
	var maxWordID int
	database.DB.Table("words").Select("max(id)").Find(&maxWordID)
	maxWordID++

	database.DB.Create(&models.Word{
		ID:        maxWordID,
		Word:      wordToAdd.Term,
		IsCreated: true,
	})

	// TODO: replace with autoincrement
	var maxDefinitionID int
	database.DB.Table("definitions").Select("max(id)").Find(&maxDefinitionID)
	maxDefinitionID++

	database.DB.Create(&models.Definition{
		ID:         maxDefinitionID,
		Definition: wordToAdd.Definition,
		WordID:     maxWordID,
	})

	collectionWord := models.CollectionWord{
		CollectionID: collection.ID,
		WordID:       maxWordID,
	}

	database.DB.Create(&collectionWord)

	database.DB.Create(&models.Priority{
		UserID:           user.ID,
		CollectionID:     collection.ID,
		CollectionWordID: collectionWord.ID,
		Priority:         1,
	})

}
