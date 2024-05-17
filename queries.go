package usermod

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TODO add support for activated or confirmed user
func NewUser(name, email string, password []byte) *User {
	return &User{
		Name:      name,
		Email:     email,
		Password:  password,
		IsDeleted: false,
	}
}

func NewUserWithPhoneNumber(name, email string, password []byte, phone string) *User {
	u := NewUser(name, email, password)
	u.PhoneNumber = phone
	return u
}

func (u *User) Insert(db *gorm.DB) error {
	result := db.Create(&u)
	return result.Error
}

func (u *User) ChangePassword(db *gorm.DB, password []byte) error {
	cryptpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	res := db.Update("password", cryptpass)
	return res.Error
}

func AuthenticateByEmail(db *gorm.DB, email, password []byte) (*User, error) {
	u := User{}
	res := db.Where("email = ?", email).First(&u)
	if res.Error != nil {
		return nil, res.Error
	}
	err := u.ValidatePassword(db, password)
	return &u, err
}

func AuthenticateByUID(db *gorm.DB, id, password []byte) (*User, error) {
	u := User{}
	res := db.Where("id = ?", id).First(&u)
	if res.Error != nil {
		return nil, res.Error
	}
	err := u.ValidatePassword(db, password)
	return &u, err
}

func (u *User) ValidatePassword(db *gorm.DB, password []byte) error {
	err := bcrypt.CompareHashAndPassword(u.Password, password)
	return err
}

func Update(db *gorm.DB, id string, name, email, phone string) error {

	u := User{Name: name, Email: email, PhoneNumber: phone}

	res := db.Model(&User{}).Where("id = ?", id).Updates(u)

	return res.Error
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
