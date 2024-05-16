package usermod

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID          uuid.UUID
	Name        []byte
	Email       []byte
	Password    []byte
	PhoneNumber []byte `json:"phone" gorm:"default:null"`
	IsDeleted   bool   `json:"-" gorm:"default:false"`
}

type UserAddress struct {
	gorm.Model
	User       User
	Street     []byte `gorm:"default:null"`
	Steet2     []byte `gorm:"default:null"`
	City       []byte `gorm:"default:null"`
	Region     []byte `gorm:"default:null"`
	Country    []byte
	PostalCode []byte `gorm:"default:null"`
}
