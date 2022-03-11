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
