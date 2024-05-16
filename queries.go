package usermod

import (
	"gorm.io/gorm"
)

func NewUser(name, email, password []byte) *User {
	return &User{
		Name:      name,
		Email:     email,
		Password:  password,
		IsDeleted: false,
	}
}

func NewUserWithPhoneNumber(name, email, password, phone []byte) *User {
	u := NewUser(name, email, password)
	u.PhoneNumber = phone
	return u
}

func (u *User) Insert(db *gorm.DB) error {
	result := db.Create(&u)
	return result.Error
}

func (u *User) ChangePassword(db *gorm.DB, password []byte) error {
	return nil
}

func (u *User) Update(db *gorm.DB, name, email, phone []byte) error {
	if len(name) > 0 {
		u.Name = name
	}
	if len(email) > 0 {
		u.Email = email
	}
	if len(phone) > 0 {
		u.PhoneNumber = phone
	}
	return nil
}

func GetOne(db *gorm.DB, id string) (*User, error) {
	u := User{}
	res := db.First(&u, id)
	return &u, res.Error
}

func DeleteByUID(db *gorm.DB, id string) error {
	res := db.Where("id = ?", id).Delete(id)
	return res.Error
}

func SoftDeleteByUID(db *gorm.DB, id string) error {
	res := db.Where("id = ?", id).Update("is_deleted", true)
	return res.Error
}
