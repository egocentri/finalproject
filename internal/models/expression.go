package models

import "gorm.io/gorm"

type Expression struct {
    gorm.Model
    UserID     uint   `gorm:"index" json:"user_id"`
    Expression string `gorm:"type:text" json:"expression"`
    Result     string `gorm:"type:text" json:"result"`
}
