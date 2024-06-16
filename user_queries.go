package usermod

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Password    []byte    `json:"-"`
	PhoneNumber string    `json:"phone"`
	IsActivated bool      `json:"-"`
	IsDeleted   bool      `json:"-"`
	db          *sql.DB
}

var userTblName = "users"

var userTblSQL = fmt.Sprintf(`CREATE TABLE %s (
	id UUID PRIMARY KEY,
	name VARCHAR(255),
	email VARCHAR(255),
	password TEXT,
	phone_number VARCHAR(255),
	is_activated BOOLEAN DEFAULT FALSE,
	is_deleted BOOLEAN DEFAULT FALSE
);`, userTblName)

func (u *User) scanInto(row *sql.Row) error {
	return row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.PhoneNumber, &u.IsActivated, &u.IsDeleted)
}

func (u *User) CreateTable() error {
	_, err := u.db.Exec(userTblSQL)
	return err
}

func (u *User) TableName() string {
	return userTblName
}

func NewUser(db *sql.DB) *User {
	return &User{db: db}
}

func NewUserWithDetails(db *sql.DB, name, email string, password []byte) *User {
	return &User{
		db:          db,
		ID:          uuid.New(),
		Name:        name,
		Email:       email,
		Password:    password,
		IsActivated: false,
		IsDeleted:   false,
	}
}

func EncryptPassword(pw []byte) []byte {
	cryptpass, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return cryptpass
}

func NewUserWithPhoneNumber(db *sql.DB, name, email string, password []byte, phone string) *User {
	u := NewUserWithDetails(db, name, email, password)
	u.PhoneNumber = phone
	return u
}

func (u *User) Insert() error {

	query := fmt.Sprintf("INSERT INTO %s VALUES ($1, $2, $3, $4, $5, $6, $7)", u.TableName())

	passwd := EncryptPassword(u.Password)

	stmt, err := u.db.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(u.ID.String(), u.Name, u.Email, passwd, u.PhoneNumber, false, false)
	if err != nil {
		return err
	}
	u.Password = passwd
	return nil
}

func Activate(db *sql.DB, id string) error {
	u := User{db: db}
	query := fmt.Sprintf(
		`UPDATE %s SET is_activated=true WHERE id = $1`, u.TableName())
	stmt, err := u.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(id)
	if err != nil {
		u.IsActivated = true
	}
	return err
}

func (u *User) Deactivate() error {
	query := fmt.Sprintf(
		`UPDATE %s SET is_activated=$1 WHERE email = $2 AND id = $3`, u.TableName())
	stmt, err := u.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(false, u.Email, u.ID.String())
	if err != nil {
		u.IsActivated = false
	}
	return err

}

func (u *User) ChangePassword(password []byte) error {
	cryptpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("UPDATE %s SET password = $1 WHERE id = $2", u.TableName())
	stmt, err := u.db.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(cryptpass, u.ID.String())
	if err != nil {
		return err
	}
	u.Password = cryptpass
	return nil
}

func AuthenticateByEmail(db *sql.DB, email string, password []byte) (*User, error) {
	u := User{db: db}
	query := fmt.Sprintf("SELECT * FROM %s WHERE email = $1 AND is_activated = $2", u.TableName())
	stmt, err := db.Prepare(query)
	if err != nil {
		return &u, err
	}
	res := stmt.QueryRow(email, true)
	if res.Err() != nil {
		return &u, res.Err()
	}

	err = u.scanInto(res)
	if err != nil {
		return &User{}, err
	}

	err = u.validatePassword(password)
	if err != nil {
		return &User{}, err
	}
	return &u, err

}

func AuthenticateByUID(db *sql.DB, id string, password []byte) (*User, error) {
	u := User{db: db}
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND is_activated = $2", u.TableName())
	stmt, err := db.Prepare(query)
	if err != nil {
		return &u, err
	}
	res := stmt.QueryRow(id, true)
	if res.Err() != nil {
		return &u, res.Err()
	}

	err = u.scanInto(res)
	if err != nil {
		return &User{}, err
	}

	err = u.validatePassword(password)
	if err != nil {
		return &User{}, err
	}
	return &u, err

}

func (u *User) validatePassword(password []byte) error {
	err := bcrypt.CompareHashAndPassword(u.Password, password)
	return err
}

// Update updates the user's details, always resetting the name,
// email, and phone_number.
func (u *User) Update(name, email, phone_number string) error {

	if name == "" {
		name = u.Name
	}
	if email == "" {
		email = u.Email
	}
	if phone_number == "" {
		phone_number = u.PhoneNumber
	}

	query := fmt.Sprintf("UPDATE %s SET ", u.TableName())
	query += " name = $1, email = $2, phone_number = $3 "
	query += "WHERE id = $4"

	stmt, err := u.db.Prepare(query)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(name, email, phone_number, u.ID.String())
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("no rows were updated")
	}

	u.Name = name
	u.Email = email
	u.PhoneNumber = phone_number
	return nil

}

func GetUserByID(db *sql.DB, id string) (*User, error) {
	u := User{db: db}
	query := fmt.Sprintf("SELECT * from %s where id = $1", u.TableName())
	stmt, err := db.Prepare(query)
	if err != nil {
		return &u, err
	}

	res := stmt.QueryRow(id)
	if res.Err() != nil {
		return &u, res.Err()
	}
	return &u, u.scanInto(res)
}

func GetActiveUserByEmail(db *sql.DB, email string) (*User, error) {
	u := User{db: db}
	query := fmt.Sprintf("SELECT * from %s where is_activated = $1 AND email = $2", u.TableName())
	stmt, err := u.db.Prepare(query)
	if err != nil {
		return &u, err
	}

	res := stmt.QueryRow(true, email)
	if res.Err() != nil {
		return &u, res.Err()
	}
	return &u, u.scanInto(res)

}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	u := User{db: db}
	query := fmt.Sprintf("SELECT * from %s where email = $1", u.TableName())
	stmt, err := u.db.Prepare(query)
	if err != nil {
		return &u, err
	}

	res := stmt.QueryRow(email)
	if res.Err() != nil {
		return &u, res.Err()
	}
	return &u, u.scanInto(res)

}

func (u *User) DeleteByUID(id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", u.TableName())
	stmt, err := u.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return err
}

func (u *User) SoftDeleteByUID(id string) error {
	query := fmt.Sprintf("UPDATE %s set is_deleted = $1 where id = $2", u.TableName())

	stmt, err := u.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(true, id)
	if err != nil {
		return err
	}

	return err
}
