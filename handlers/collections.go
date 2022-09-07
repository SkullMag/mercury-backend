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
	database.DB.Select("name").Where("user_id = ? and name = ?", user.ID, strings.ToLower(vars["name"])).Find(&collection)
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
	var collectionWords []models.CollectionWord
	var priorities []models.Priority
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

	database.DB.Where("collection_id = ?", collection.ID).Find(&collectionWords)
	database.DB.Where("collection_id = ? and user_id = ?", collection.ID, user.ID).Find(&priorities)
	database.DB.Delete(&collectionWords)
	database.DB.Delete(&priorities)
	database.DB.Delete(&collection)
}

func GetCollections(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var user models.User

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	database.DB.Preload("Collections.Words.Word").Preload("Collections.User").Where("username = ?", vars["username"]).Find(&user)
	if user.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "User was not found"}`)
		return
	}

	// if user.Username != vars["username"] && !user.IsSubscribed {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	fmt.Fprint(w, `{"error": "Subscribe to see another users collections"}`)
	// 	return
	// }


    for i := 0; i < len(user.Collections); i++ {
        var favourite models.Favourite
        res := database.DB.Where("user_id = ? AND collection_id = ?", user.ID, user.Collections[i].ID).Find(&favourite)
        user.Collections[i].IsFavourite = res.RowsAffected > 0
        if word, ok := vars["word"]; ok {
            status := false
            for _, colWord := range user.Collections[i].Words {
                if word == colWord.Word.Word {
                    status = true
                    break
                }
            }
            user.Collections[i].ContainsWord = status
       }
    }

	response, _ := json.Marshal(&user.Collections)
	fmt.Fprint(w, string(response))

}

func GetAllCollections(w http.ResponseWriter, req *http.Request) {
    vars := mux.Vars(req)
    var user models.User
	var collections []models.Collection

	database.DB.Preload("User").Preload("Words").Limit(100).Find(&collections)
    if token, ok := vars["token"]; ok {
        if !utils.AuthenticateToken(&w, req, &user, token) {
            return
        }
        for i := 0; i < len(collections); i++ {
            var favourite models.Favourite
            res := database.DB.Where("user_id = ? AND collection_id = ?", user.ID, collections[i].ID).Find(&favourite)
            collections[i].IsFavourite = res.RowsAffected > 0
        }
    }
	response, _ := json.Marshal(&collections)
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

	res := database.DB.Preload("Words.Word.Definitions").Where("name = ? and user_id = ?", strings.ToLower(vars["collectionName"]), requestedUser.ID).Find(&collection)
	if res.Error != nil {
		fmt.Println(res.Error)
	}
	if collection.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Collection was not found"}`)
		return
	}

	// if user.ID != collection.UserID && !user.IsSubscribed {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	fmt.Fprint(w, `{"error": "Subscribe to see another users collections"}`)
	// 	return
	// }
	wordsToReturn := utils.GenerateWordsJSON(collection.Words, user)

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

	res := database.DB.Where("name = ? and user_id = ?", strings.ToLower(vars["collectionName"]), user.ID).Find(&collection)
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
		// Newly added words can have priority 0 | 1
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
	var priorities []models.Priority

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

	database.DB.Where("collection_word_id = ? and collection_id = ? and user_id = ?", collectionWord.ID, collection.ID, user.ID).Find(&priorities)

	database.DB.Delete(&priorities)
	database.DB.Delete(&collectionWord)
}

func RenameCollection(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	var user models.User
	var collection models.Collection

	if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
		return
	}

	if res := database.DB.Where("name = ?", vars["oldName"]).Find(&collection); res.RowsAffected == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "Collection was not found"}`)
		return
	}

	collection.Name = vars["newName"]
	database.DB.Save(&collection)
}

func AddCollectionToFavourites(w http.ResponseWriter, req *http.Request) {
    vars := mux.Vars(req)

    var user models.User
    var collection models.Collection
    var favourite models.Favourite
    
    if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
        return
    }

    if res := database.DB.Where("id = ?", vars["collection_id"]).Find(&collection); res.RowsAffected == 0 {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, `{"error": "Collection doesn't exist"}`)
        return
    }

    if res := database.DB.Where("collection_id = ? AND user_id = ?", collection.ID, user.ID).Find(&favourite); res.RowsAffected > 0 {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, `{"error": "Collection is already in favourites"}`)
        return
    }

    favourite.UserID = user.ID
    favourite.CollectionID = collection.ID

    database.DB.Create(&favourite)
}

func RemoveCollectionFromFavourites(w http.ResponseWriter, req *http.Request) {
    vars := mux.Vars(req)

    var user models.User
    var favourite models.Favourite

    if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
        return
    }

    res := database.DB.Where("user_id = ? AND collection_id = ?", user.ID, vars["collection_id"]).Find(&favourite)
    if res.RowsAffected == 0 {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, `{"error": "Collection is not in favourites"}`)
        return
    }
    
    database.DB.Delete(&favourite)

}

func GetFavouriteCollections(w http.ResponseWriter, req *http.Request) {
    vars := mux.Vars(req)
    
    var user models.User
    var favourites []models.Favourite
    var collections []models.Collection = make([]models.Collection, 0)

    if !utils.AuthenticateToken(&w, req, &user, vars["token"]) {
        return
    }

	database.DB.Preload("Collection.Words.Word").Preload("Collection.User").Where("user_id = ?", user.ID).Find(&favourites)

    
    for _, favourite := range favourites {
        favourite.Collection.IsFavourite = true
        collections = append(collections, favourite.Collection)
    }

    response, _ := json.Marshal(collections)
    fmt.Fprint(w, string(response))
}
