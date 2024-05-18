package usermod

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Password    []byte    `json:"password"`
	PhoneNumber string    `json:"phone" gorm:"default:null"`
	IsDeleted   bool      `json:"-" gorm:"default:false"`
}

type UserAddress struct {
	gorm.Model
	User       User
	Street     string `gorm:"default:null"`
	Steet2     string `gorm:"default:null"`
	City       string `gorm:"default:null"`
	Region     string `gorm:"default:null"`
	Country    string
	PostalCode string `gorm:"default:null"`
}
