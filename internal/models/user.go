package models

import "gorm.io/gorm"

type User struct {
    gorm.Model
    Login        string `gorm:"unique;not null" json:"login"`
    PasswordHash string `gorm:"not null"     json:"-"`
}
