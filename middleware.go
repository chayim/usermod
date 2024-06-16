package usermod

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type CTXvar string

const (
	CTX_UID_KEY  CTXvar = "uid"
	CTX_USER_KEY CTXvar = "user"
)

func BasicAuth(db *sql.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			uobj, err := AuthenticateByEmail(db, user, []byte(pass))
			if err != nil {
				jsonError(w, err, http.StatusForbidden)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), CTX_USER_KEY, uobj))
			r = r.WithContext(context.WithValue(r.Context(), CTX_UID_KEY, uobj.ID.String()))
			next.ServeHTTP(w, r)
		})
	}
}

// TODO implement jwt auth middleware
func JWTTokenAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokens := strings.Split(tokenString, " ")
		if len(tokens) != 2 || tokens[0] != "Bearer" {
			http.Error(w, "Unauthorized", http.StatusBadRequest)
			return
		}

		token, err := jwt.ParseWithClaims(tokens[1], &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*Claims)
		r = r.WithContext(context.WithValue(r.Context(), CTX_UID_KEY, claims.UserID))
		next.ServeHTTP(w, r)
	})
}

// TODO implement session auth middleware
func SessionAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}
