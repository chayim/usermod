package usermod

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TODO add support for activated or confirmed user
func NewUser(name, email string, password []byte) *User {
	cryptpass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &User{
		ID:          uuid.New(),
		Name:        name,
		Email:       email,
		Password:    cryptpass,
		IsActivated: false,
		IsDeleted:   false,
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

func ActivateUser(db *gorm.DB, email string) error {
	res := db.Model(&User{}).Where("email = ?", email).Update("is_activated", true)
	return res.Error
}

func (u *User) ChangePassword(db *gorm.DB, password []byte) error {
	cryptpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = cryptpass
	res := db.Save(&u)
	return res.Error
}

func AuthenticateByEmail(db *gorm.DB, email string, password []byte) (*User, error) {
	u := User{}
	res := db.Model(&User{}).Where("email = ?", email).First(&u)
	if res.Error != nil {
		return nil, res.Error
	}
	err := u.validatePassword(password)
	if err != nil {
		return nil, err
	}
	return &u, err
}

func AuthenticateByUID(db *gorm.DB, id string, password []byte) (*User, error) {
	u := User{}
	res := db.Model(&User{}).Where("id = ?", id).First(&u)
	if res.Error != nil {
		return nil, res.Error
	}
	err := u.validatePassword(password)
	if err != nil {
		return nil, err
	}
	return &u, err
}

func (u *User) validatePassword(password []byte) error {
	err := bcrypt.CompareHashAndPassword(u.Password, password)
	return err
}

func Update(db *gorm.DB, id string, name, email, phone string) error {

	// only non empty fields will be updated
	u := User{Name: name, Email: email, PhoneNumber: phone}

	res := db.Model(&User{}).Where("id = ?", id).Updates(u)

	return res.Error
}

func GetOne(db *gorm.DB, id string) (*User, error) {
	u := User{}
	uid, _ := uuid.Parse(id)
	res := db.First(&u, uid)
	return &u, res.Error
}

func DeleteByUID(db *gorm.DB, id string) error {
	res := db.Model(&User{}).Where("id = ?", id).Delete(id)
	return res.Error
}

func SoftDeleteByUID(db *gorm.DB, id string) error {
	res := db.Model(&User{}).Where("id = ?", id).Update("is_deleted", true)
	return res.Error
}
