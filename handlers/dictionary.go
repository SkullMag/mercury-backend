package handlers

import (
	"encoding/json"
	"fmt"
	"mercury/database"
	"mercury/models"
	"mercury/utils"
	"net/http"

	"github.com/gorilla/mux"
)

func GetDefinition(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	vars := mux.Vars(req)

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

func CreateCollection(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	var user models.User
	var collection models.Collection
	vars := mux.Vars(req)

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	if len(vars["name"]) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Collection name should be at least 3 characters long"}`)
		return
	}

	// Better to handle this with the unique constraint
	// to reduce number of database requests
	database.DB.Select("name").Where("user_id = ? and name = ?", user.ID, vars["name"]).Find(&collection)
	if collection.Name != "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "You've already created a collection with this name"}`)
		return
	}

	database.DB.Create(&models.Collection{
		Name:   vars["name"],
		UserID: user.ID,
	})

}

func GetCollections(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	vars := mux.Vars(req)
	var user models.User

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	database.DB.Preload("Collections.Words").Where("username = ?", vars["username"]).Find(&user)
	if user.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "User was not found"}`)
		return
	}

	if user.Username != vars["username"] && !user.IsSubscribed {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Subscribe to see another users collections"}`)
		return
	}

	response, _ := json.Marshal(&user.Collections)
	fmt.Fprint(w, string(response))

}

func GetCollectionWords(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	vars := mux.Vars(req)
	var user models.User
	var requestedUser models.User
	var collection models.Collection

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	database.DB.Where("username = ?", vars["createdByUsername"]).Find(&requestedUser)
	if requestedUser.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "User was not found"}`)
		return
	}

	database.DB.Preload("Words.Word.Definitions").Where("name = ? and user_id = ?", vars["collectionName"], requestedUser.ID).Find(&collection)
	if collection.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Collection was not found"}`)
		return
	}

	if user.ID != collection.UserID && !user.IsSubscribed {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Subscribe to see another users collections"}`)
		return
	}

	response, _ := json.Marshal(&collection.Words)
	fmt.Fprint(w, string(response))
}

func AddWordsToCollection(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	vars := mux.Vars(req)
	var user models.User
	var collection models.Collection
	var words []string

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	database.DB.Preload("Words.Word").Where("name = ? and user_id = ?", vars["collectionName"], user.ID).Find(&collection)
	if collection.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Collection was not found"}`)
		return
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&words); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"status": "JSON encoding error"}`)
		return
	}

	for _, word := range words {
		go func(word string) {
			var dbWord models.Word
			var collectionWord models.CollectionWord
			var priority models.Priority
			response := database.DB.Where("word = ?", word).Find(&dbWord)
			if response.RowsAffected > 0 {
				collectionWord.CollectionID = collection.ID
				collectionWord.WordID = dbWord.ID

				priority.UserID = user.ID
				priority.CollectionID = collection.ID
				priority.WordID = dbWord.ID
				priority.Priority = 1
				if database.DB.Create(&collectionWord).Error != nil {
					return
				}
				database.DB.Create(&priority)
			}
		}(word)
	}
}
