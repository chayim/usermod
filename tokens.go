package usermod

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	Email  string `json:"email"`
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func CreateToken(u *User) (string, error) {
	mins, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION"))
	if err != nil || mins <= 0 {
		mins = 15
	}

	// Token expires in 15 minutes
	expirationTime := time.Now().Add(time.Duration(mins) * time.Minute)
	claims := &Claims{
		Email:  u.Email,
		UserID: u.ID.String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
