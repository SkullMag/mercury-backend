package main

import (
	"mercury/handlers"
	"net/http"
)

func main() {
	http.HandleFunc("/api/register", handlers.Register)
	http.HandleFunc("/api/login", handlers.Login)
	http.HandleFunc("/api/checkToken", handlers.CheckToken)

	http.ListenAndServe(":8080", nil)
}
