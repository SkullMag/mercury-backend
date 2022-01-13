package handlers


import (
    "net/http"
    "fmt"
    "encoding/json"
    "encoding/hex"
    "crypto/sha512"
    "mercury/database"
    "mercury/models"
    "mercury/utils"
)


func Register(w http.ResponseWriter, req *http.Request) {
    errorResponse := map[string]string{"error": ""}
    if req.Method != "POST" {
        w.WriteHeader(405)
        return
    }
    var user models.User
    decoder := json.NewDecoder(req.Body)
    err := decoder.Decode(&user)
    if err != nil || user.Username == "" || user.Password == "" {
        w.WriteHeader(400)
        errorResponse["error"] = "Username or password was not provided"
        response, _ := json.Marshal(errorResponse)
        fmt.Fprintf(w, string(response))
        return
    }
    token, _ := utils.GenerateRandomStringURLSafe(32)
    salt, _ := utils.GenerateRandomStringURLSafe(32)
    hasher := sha512.New()
    hasher.Write([]byte(user.Password + salt))
    user.Password = hex.EncodeToString(hasher.Sum(nil))
    user.Token = token
    user.Salt = salt
    result := database.DB.Create(&user)
    if result.Error != nil {
        w.WriteHeader(400)
        errorResponse["error"] = "Username is not unique"
        response, _ := json.Marshal(errorResponse)
        fmt.Fprintf(w, string(response))
        return
    }
}


func Login(w http.ResponseWriter, req* http.Request) {
    var user models.User
    var dbUser models.User
    decoder := json.NewDecoder(req.Body)
    err := decoder.Decode(&user)
    if err != nil {
        w.WriteHeader(400)
    }
    database.DB.Where("username = ?", user.Username).First(&dbUser)
    if dbUser.Salt == "" {
        w.WriteHeader(400)
        return
    }
    hasher := sha512.New()
    hasher.Write([]byte(user.Password + dbUser.Salt))
    if hex.EncodeToString(hasher.Sum(nil)) == dbUser.Password {
        response, _ := json.Marshal(map[string]string{"token": dbUser.Token})
        fmt.Fprintf(w, string(response))
    }
}