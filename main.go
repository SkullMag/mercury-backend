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
	api.HandleFunc("/signup", handlers.SignUp).Methods("POST")
	api.HandleFunc("/login", handlers.Login).Methods("POST")
	api.HandleFunc("/getUserData/{token}", handlers.GetUserData).Methods("GET")
	api.HandleFunc("/getUserProfilePicture/{username}", handlers.GetUserProfilePicture).Methods("GET")
	api.HandleFunc("/definition/{word}", handlers.GetDefinition).Methods("GET")
	api.HandleFunc("/requestVerificationCode/{username}/{email}", handlers.RequestVerificationCode).Methods("GET")
	api.HandleFunc("/createCollection/{token}/{name}", handlers.CreateCollection).Methods("POST")
	api.HandleFunc("/deleteCollection/{token}/{collectionName}", handlers.DeleteCollection).Methods("POST")
	api.HandleFunc("/getCollections/{token}/{username}", handlers.GetCollections).Methods("GET")
	api.HandleFunc("/getCollectionWords/{token}/{createdByUsername}/{collectionName}", handlers.GetCollectionWords).Methods("GET")
	api.HandleFunc("/addWordToCollection/{token}/{collectionName}/{word}", handlers.AddWordToCollection).Methods("POST")
	api.HandleFunc("/deleteCollectionWord/{token}/{collectionName}/{word}", handlers.DeleteCollectionWord).Methods("POST")
	api.HandleFunc("/learnWords/{token}/{collectionName}", handlers.LearnWords).Methods("POST")

	if val, ok := os.LookupEnv("MERCURY_PRODUCTION"); ok && val == "1" {
		router.PathPrefix("/").Handler(http.FileServer(http.Dir("build")))
	}

	origins := gorillaHandlers.AllowedOrigins([]string{"*"})
	http.ListenAndServe(":8080", gorillaHandlers.CORS(origins)(router))
}
