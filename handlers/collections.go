package handlers

import (
	"encoding/json"
	"fmt"
	"mercury/database"
	"mercury/models"
	"mercury/utils"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func CreateCollection(w http.ResponseWriter, req *http.Request) {
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
		Name:   strings.ToLower(vars["name"]),
		UserID: user.ID,
	})

}

func DeleteCollection(w http.ResponseWriter, req *http.Request) {
	var user models.User
	var collection models.Collection
	vars := mux.Vars(req)

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	result := database.DB.Where("name = ? and user_id = ?", strings.ToLower(vars["collectionName"]), user.ID).Find(&collection)
	if result.RowsAffected == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "No collection with specified name was found"}`)
		return
	}
	database.DB.Delete(&collection)
}

func GetCollections(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var user models.User

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	database.DB.Preload("Collections.Words").Preload("Collections.User").Where("username = ?", vars["username"]).Find(&user)
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

	var wordsToReturn []map[string]interface{}
	for _, word := range collection.Words {
		definitions := make(map[string][]map[string]string)

		for _, element := range word.Word.Definitions {
			definitions[element.PartOfSpeech] = append(definitions[element.PartOfSpeech], map[string]string{
				"definition": element.Definition,
				"example":    element.Example,
			})
		}
		result := make(map[string]interface{})
		result["word"] = word.Word.Word
		result["definitions"] = definitions
		result["phonetics"] = word.Word.Phonetics
		wordsToReturn = append(wordsToReturn, result)
	}

	response, _ := json.Marshal(wordsToReturn)
	fmt.Fprint(w, string(response))
}

func AddWordToCollection(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var user models.User
	var collection models.Collection

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

	// decoder := json.NewDecoder(req.Body)
	// if err := decoder.Decode(&words); err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	fmt.Fprint(w, `{"status": "JSON encoding error"}`)
	// 	return
	// }

	var dbWord models.Word
	var collectionWord models.CollectionWord
	var priority models.Priority
	response := database.DB.Where("word = ?", strings.ToLower(vars["word"])).Find(&dbWord)
	if response.RowsAffected > 0 {
		collectionWord.CollectionID = collection.ID
		collectionWord.WordID = dbWord.ID
		if err := database.DB.Create(&collectionWord).Error; err != nil {
			fmt.Println(err.Error())
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

func DeleteCollectionWord(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var user models.User
	var word models.Word
	var collection models.Collection
	var collectionWord models.CollectionWord

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	if res := database.DB.Where("word = ?", strings.ToLower(vars["word"])).Find(&word); res.RowsAffected == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Word was not found"}`)
		return
	}

	if res := database.DB.Where("name = ? and user_id = ?", strings.ToLower(vars["collectionName"]), user.ID).Find(&collection); res.RowsAffected == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Collection was not found"}`)
		return
	}

	if res := database.DB.Where("word_id = ? and collection_id = ?", word.ID, collection.ID).Find(&collectionWord); res.RowsAffected == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Word is not in collection"}`)
		return
	}

	database.DB.Delete(&collectionWord)
}

// func DeleteWordsFromCollection(w http.ResponseWriter, req *http.Request) {
// 	vars := mux.Vars(req)
// 	var user models.User
// 	var collection models.Collection

// 	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
// 		return
// 	}

// 	if database.DB.Where("name = ? and user_id = ?", vars["collectionName"], user.ID).Find(&collection).RowsAffected == 0 {
// 		w.WriteHeader(http.StatusBadRequest)
// 		fmt.Fprint(w, `{"error": "No collection with specified name was found"}`)
// 		return
// 	}

// 	var words []string
// 	decoder := json.NewDecoder(req.Body)
// 	if decoder.Decode(&words) != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		fmt.Fprint(w, `{"error": "JSON decoding error"}`)
// 		return
// 	}

// 	for _, word := range words {
// 		var wordID int
// 		var collectionWord models.CollectionWord

// 		database.DB.Table("words").Select("id").Where("word = ?", word).Find(&wordID)
// 		database.DB.Clauses(clause.Returning{}).Where("word_id = ? and collection_id = ?", wordID, collection.ID).Delete(&collectionWord)
// 		database.DB.Where("collection_id = ? and collection_word_id = ?", collection.ID, collectionWord.ID).Delete(models.Priority{})
// 	}

// }

func GetWordsToLearn(w http.ResponseWriter, req *http.Request) {
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
