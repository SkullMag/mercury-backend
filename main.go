package main


import (
    "net/http"
    "fmt"
    "mercury/handlers"
    "mercury/database"
    "mercury/models"
)


func main() {
    var user models.User
    database.DB.First(&user)
    fmt.Println(user.Username)

    http.HandleFunc("/api/register", handlers.Register)

    http.ListenAndServe(":8080", nil)
}
