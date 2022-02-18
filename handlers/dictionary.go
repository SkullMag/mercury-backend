package handlers

import (
	"encoding/json"
	"fmt"
	"mercury/database"
	"mercury/models"
	"mercury/utils"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm/clause"
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

	res := database.DB.Preload("Words.Word.Definitions").Where("name = ? and user_id = ?", vars["collectionName"], requestedUser.ID).Find(&collection)
	if res.Error != nil {
		fmt.Println(res.Error)
	}
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

	for i := 0; i < len(collection.Words); i++ {
		var priority models.Priority
		res := database.DB.Table("priorities").Select("priority").Where("user_id = ? and collection_id = ? and collection_word_id = ?", user.ID, collection.ID, collection.Words[i].ID).Find(&priority)
		if res.Error == nil {
			collection.Words[i].Priority = priority.Priority
		}
	}

	response, _ := json.Marshal(&collection.Words)
	fmt.Fprint(w, string(response))
}

func AddWordToCollection(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	vars := mux.Vars(req)
	var user models.User
	var collection models.Collection
	var words []string

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	res := database.DB.Where("name = ? and user_id = ?", vars["collectionName"], user.ID).Find(&collection)
	if res.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(res.Error)
		return
	}
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

	var dbWord models.Word
	var collectionWord models.CollectionWord
	var priority models.Priority
	response := database.DB.Where("word = ?", vars["word"]).Find(&dbWord)
	if response.RowsAffected > 0 {
		collectionWord.CollectionID = collection.ID
		collectionWord.WordID = dbWord.ID
		if database.DB.Create(&collectionWord).Error != nil {
			return
		}
		priority.UserID = user.ID
		priority.CollectionID = collection.ID
		priority.CollectionWordID = collectionWord.ID
		priority.Priority = 1
		database.DB.Create(&priority)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Word was not found in dictionary"}`)
	}
}

func DeleteWordsFromCollection(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	vars := mux.Vars(req)
	var user models.User
	var collection models.Collection

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	if database.DB.Where("name = ? and user_id = ?", vars["collectionName"], user.ID).Find(&collection).RowsAffected == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "No collection with specified name was found"}`)
		return
	}

	var words []string
	decoder := json.NewDecoder(req.Body)
	if decoder.Decode(&words) != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "JSON decoding error"}`)
		return
	}

	for _, word := range words {
		var wordID int
		var collectionWord models.CollectionWord

		database.DB.Table("words").Select("id").Where("word = ?", word).Find(&wordID)
		database.DB.Clauses(clause.Returning{}).Where("word_id = ? and collection_id = ?", wordID, collection.ID).Delete(&collectionWord)
		database.DB.Where("collection_id = ? and collection_word_id = ?", collection.ID, collectionWord.ID).Delete(models.Priority{})
	}

}

func GetWordsToLearn(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	vars := mux.Vars(req)
	var user models.User
	var requestedUser models.User
	var collection models.Collection

	if status := utils.AuthenticateToken(&w, req, &user, vars["token"]); !status {
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

	var result []string
	database.DB.Raw("select w.word from priorities p left join collection_words cw on p.collection_word_id = cw.id left join words w on w.id = cw.word_id order by p.priority limit 20").Scan(&result)
	response, _ := json.Marshal(result)
	fmt.Fprint(w, string(response))
}
