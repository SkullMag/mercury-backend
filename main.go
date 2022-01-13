package main


import (
    "net/http"
    "mercury/handlers"
)


func main() {
    http.HandleFunc("/api/register", handlers.Register)
    http.HandleFunc("/api/login", handlers.Login)

    http.ListenAndServe(":8080", nil)
}
