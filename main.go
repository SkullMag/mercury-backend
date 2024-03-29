package main

import (
	"mercury/handlers"
	"net/http"
	"os"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	api := router.PathPrefix("/api").Subrouter()

	// GET
	api.HandleFunc("/getUserData/{token}", handlers.GetUserData).Methods("GET")
	api.HandleFunc("/getUserData/{token}/{username}", handlers.GetUserDataByUsername).Methods("GET")
	api.HandleFunc("/getUserProfilePicture/{username}", handlers.GetUserProfilePicture).Methods("GET")
	api.HandleFunc("/getFavouriteCollections/{token}", handlers.GetFavouriteCollections).Methods("GET")
	api.HandleFunc("/getCollections", handlers.GetAllCollections).Methods("GET")
	api.HandleFunc("/getCollections/{token}", handlers.GetAllCollections).Methods("GET")
	api.HandleFunc("/getCollections/{token}/{username}", handlers.GetCollections).Methods("GET")
	api.HandleFunc("/getCollections/{token}/{username}/{word}", handlers.GetCollections).Methods("GET")
	api.HandleFunc("/getUserStats/{username}", handlers.GetUserStats).Methods("GET")
	api.HandleFunc("/getCollectionWords/{token}/{createdByUsername}/{collectionName}", handlers.GetCollectionWords).Methods("GET")
	api.HandleFunc("/definition/{word}", handlers.GetDefinition).Methods("GET")
	api.HandleFunc("/requestVerificationCode/{username}/{email}", handlers.RequestVerificationCode).Methods("GET")
	api.HandleFunc("/getSubscriptionStatus/{username}", handlers.IsSubscribed).Methods("GET")

	// POST
	api.HandleFunc("/signup", handlers.SignUp).Methods("POST")
	api.HandleFunc("/login", handlers.Login).Methods("POST")
	api.HandleFunc("/createCollection/{token}/{name}", handlers.CreateCollection).Methods("POST")
	api.HandleFunc("/addWordToCollection/{token}/{collectionName}/{word}", handlers.AddWordToCollection).Methods("POST")
	api.HandleFunc("/learnWords/{token}/{authorUsername}/{collectionName}", handlers.LearnWords).Methods("POST")
	api.HandleFunc("/renameCollection/{token}/{oldName}/{newName}", handlers.RenameCollection).Methods("POST")
	api.HandleFunc("/addNewWord/{token}", handlers.AddWord).Methods("POST")

	api.HandleFunc("/deleteCollection/{token}/{collectionName}", handlers.DeleteCollection).Methods("POST")
	api.HandleFunc("/deleteCollectionWord/{token}/{collectionName}/{word}", handlers.DeleteCollectionWord).Methods("POST")
	api.HandleFunc("/deleteProfile/{token}", handlers.DeleteProfile).Methods("POST")

	api.HandleFunc("/addCollectionToFavourites/{token}/{collection_id}", handlers.AddCollectionToFavourites).Methods("POST")
	api.HandleFunc("/removeCollectionFromFavourites/{token}/{collection_id}", handlers.RemoveCollectionFromFavourites).Methods("POST")

	origins := gorillaHandlers.AllowedOrigins([]string{"*"})
	// http.ListenAndServeTLS(":"+os.Getenv("PORT"), "mercurydict.com.crt", "mercurydict.com.key", gorillaHandlers.CORS(origins)(router))
	http.ListenAndServe(":"+os.Getenv("PORT"), gorillaHandlers.CORS(origins)(router))
}
