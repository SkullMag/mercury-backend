package main

import (
	"mercury/handlers"
	"net/http"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/signup", handlers.SignUp).Methods("POST")
	router.HandleFunc("/api/login", handlers.Login).Methods("POST")
	router.HandleFunc("/api/getUserData/{token}", handlers.GetUserData).Methods("GET")
	router.HandleFunc("/api/getUserProfilePicture/{username}", handlers.GetUserProfilePicture).Methods("GET")
	router.HandleFunc("/api/definition/{word}", handlers.GetDefinition).Methods("GET")
	router.HandleFunc("/api/requestVerificationCode/{username}/{email}", handlers.RequestVerificationCode).Methods("GET")
	router.HandleFunc("/api/createCollection/{token}/{name}", handlers.CreateCollection).Methods("POST")
	router.HandleFunc("/api/deleteCollection/{token}/{collectionName}", handlers.DeleteCollection).Methods("POST")
	router.HandleFunc("/api/getCollections/{token}/{username}", handlers.GetCollections).Methods("GET")
	router.HandleFunc("/api/getCollectionWords/{token}/{createdByUsername}/{collectionName}", handlers.GetCollectionWords).Methods("GET")
	router.HandleFunc("/api/addWordToCollection/{token}/{collectionName}/{word}", handlers.AddWordToCollection).Methods("POST")
	router.HandleFunc("/api/deleteCollectionWord/{token}/{collectionName}/{word}", handlers.DeleteCollectionWord).Methods("POST")
	// router.HandleFunc("/api/deleteWordsFromCollection/{token}/{collectionName}", handlers.DeleteWordsFromCollection).Methods("POST")
	router.HandleFunc("/api/getWordsToLearn/{token}/{createdByUsername}/{collectionName}", handlers.GetWordsToLearn).Methods("GET")

	origins := gorillaHandlers.AllowedOrigins([]string{"*"})
	http.ListenAndServe(":8080", gorillaHandlers.CORS(origins)(router))
}
