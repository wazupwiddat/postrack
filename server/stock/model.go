package stock

import (
	"gorm.io/gorm"
)

type Stock struct {
	gorm.Model
	ID     uint   `gorm:"primary_key"`
	UserID uint   `gorm:"index"`
	Symbol string `gorm:"size:100;unique;not null"`
	Name   string `gorm:"size:100"`
}
