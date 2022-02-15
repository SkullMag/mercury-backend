package main

import (
	"mercury/handlers"
	"net/http"

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
	router.HandleFunc("/api/getCollections/{token}/{username}", handlers.GetCollections).Methods("GET")
	router.HandleFunc("/api/getCollectionWords/{token}/{username}/{name}", handlers.GetCollectionWords).Methods("GET")

	http.ListenAndServe(":8080", router)
}
