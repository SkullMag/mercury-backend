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

func CreateCollection(w http.ResponseWriter, req *http.Request) {
	utils.EnableCors(&w)

	var user models.User
	var collection models.Collection

	vars := mux.Vars(req)
	if _, ok := vars["token"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "no token provided"}`)
		return
	}
	if _, ok := vars["name"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "no collection name provided"}`)
		return
	}
	if len(vars["name"]) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Collection name should be at least 3 characters long"}`)
		return
	}

	database.DB.Select("username").Where("token = ?", vars["token"]).Find(&user)
	if user.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Invalid token"}`)
		return
	}

	collection.Name = vars["name"]

}
