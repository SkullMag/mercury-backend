package database

import (
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
)


var DB, err = gorm.Open(sqlite.Open("mercury.db"), &gorm.Config{})
