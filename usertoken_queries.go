package usermod

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Token int

const (
	ForgotPaswordToken Token = iota
	ActivationToken
)

// Antipattern, this relies on email and not the foreign eky to user
// which makes it much faster, as down the line no join needs to happen
type UserOperationToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"-"`
	Expiry    int64     `json:"-"`
	TokenType Token     `json:"tokenType"`
	Used      bool      `json:"-"`
	db        *sql.DB
}

var TokenDefaultExpiry = time.Hour * 48 // dfeault expire 48 hours
var userOpsTokenTblName = "user_ops_tokens"
var userOpsTokenTblSQL = fmt.Sprintf(`CREATE TABLE %s (
	id UUID PRIMARY KEY,
	user_id UUID,
	expiry int,
	token_type int,
	used BOOLEAN DEFAULT FALSE
);`, userOpsTokenTblName)

func NewUserOperationToken(db *sql.DB) *UserOperationToken {
	return &UserOperationToken{db: db}
}

func NewUserOperationTokenWithExpires(db *sql.DB, user uuid.UUID, tokenType Token, expires time.Time) *UserOperationToken {
	return &UserOperationToken{
		ID:        uuid.New(),
		UserID:    user,
		TokenType: tokenType,
		db:        db,
		Expiry:    expires.Unix(),
	}
}

func NewUserOperationTokenDefaultExpires(db *sql.DB, user uuid.UUID, tokenType Token) *UserOperationToken {
	return &UserOperationToken{
		ID:        uuid.New(),
		UserID:    user,
		TokenType: tokenType,
		db:        db,
		Expiry:    time.Now().Add(TokenDefaultExpiry).Unix(),
	}

}
func (u *UserOperationToken) TableName() string {
	return userOpsTokenTblName
}

func (u *UserOperationToken) scanInto(row *sql.Row) error {
	return row.Scan(&u.ID, &u.UserID, &u.Expiry, &u.TokenType, &u.Used)
}

func (u *UserOperationToken) CreateTable() error {
	_, err := u.db.Exec(userOpsTokenTblSQL)
	return err
}
func (u *UserOperationToken) Insert() error {
	query := fmt.Sprintf("INSERT INTO %s VALUES ($1, $2, $3, $4, $5)", u.TableName())

	stmt, err := u.db.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(u.ID, u.UserID.String(), u.Expiry, u.TokenType, u.Used)
	return err
}

func GetUserOperationToken(db *sql.DB, token string) (*UserOperationToken, error) {
	u := UserOperationToken{db: db}
	query := fmt.Sprintf("SELECT * from %s WHERE id = $1", u.TableName())
	stmt, err := db.Prepare(query)
	if err != nil {
		return &u, err
	}
	res := stmt.QueryRow(token)
	if res.Err() != nil {
		return &u, res.Err()
	}
	return &u, u.scanInto(res)
}

func GetTokenIfValid(db *sql.DB, uid, token string) (*UserOperationToken, error) {
	u := UserOperationToken{db: db}
	query := fmt.Sprintf(
		`SELECT * from %s
		WHERE user_id = $1 AND
		id = $2 AND
		used = $3 AND
		expiry >= $4
		`, u.TableName())
	now := time.Now().Unix()

	stmt, err := db.Prepare(query)
	if err != nil {
		return &u, err
	}
	res := stmt.QueryRow(uid, token, false, now)
	if res.Err() != nil {
		return &u, res.Err()
	}
	return &u, u.scanInto(res)
}

func MarkTokenAsUsed(db *sql.DB, tok string) (uuid.UUID, error) {
	u := UserOperationToken{db: db}
	query := fmt.Sprintf(`
		UPDATE %s SET used = $1
		WHERE
		id = $2
		RETURNING user_id`, u.TableName())

	stmt, err := u.db.Prepare(query)
	if err != nil {
		return uuid.Nil, err
	}

	res := stmt.QueryRow(true, tok)
	if res.Err() != nil {
		return uuid.Nil, res.Err()
	}
	var uid uuid.UUID
	err = res.Scan(&uid)
	return uid, err
}
