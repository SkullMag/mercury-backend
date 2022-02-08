package main

import (
	"mercury/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/register", handlers.Register).Methods("POST")
	router.HandleFunc("/api/login", handlers.Login).Methods("POST")
	router.HandleFunc("/api/checkToken/{token}", handlers.CheckToken).Methods("POST")
	router.HandleFunc("/api/getUserData/{token}", handlers.GetUserData).Methods("GET")
	router.HandleFunc("/api/getUserProfilePicture/{username}", handlers.GetUserProfilePicture).Methods("GET")
	router.HandleFunc("/api/definition/{word}", handlers.GetDefinition).Methods("GET")
	router.HandleFunc("/api/requestVerificationCode/{email}", handlers.RequestVerificationCode).Methods("GET")

	http.ListenAndServe(":8080", router)
}
