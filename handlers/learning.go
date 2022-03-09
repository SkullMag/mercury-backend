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

func LearnWords(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var user models.User

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	var words []map[string]string

	decoder := json.NewDecoder(req.Body)

	if err := decoder.Decode(&words); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Wrong words specified"}`)
		return
	}

	for _, word := range words {
		var w models.Word
		var collection models.Collection
		var collectionWord models.CollectionWord
		var priority models.Priority

		if res := database.DB.Where("word = ?", word["word"]).Find(&w); res.RowsAffected == 0 {
			continue
		}
		if res := database.DB.Where("user_id = ? and name = ?", user.ID, strings.ToLower(vars["collectionName"])).Find(&collection); res.RowsAffected == 0 {
			continue
		}
		if res := database.DB.Where("collection_id = ? and word_id = ?", collection.ID, w.ID).Find(&collectionWord); res.RowsAffected == 0 {
			continue
		}
		if res := database.DB.Where("collection_id = ? and collection_word_id = ? and user_id = ?", collection.ID, collectionWord.ID, user.ID).Find(&priority); res.RowsAffected == 0 {
			continue
		}

		if word["status"] == "true" {
			priority.Priority += 1
		} else {
			priority.Priority = 1
		}

		database.DB.Save(&priority)
	}

}
