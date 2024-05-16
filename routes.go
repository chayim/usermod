package usermod

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	// r.Get("/users", ListUsers)
	r.Get("/users/{id}", Get)
	// r.Post("/users", CreateUser)
	// r.Put("/users/{id}", UpdateUser)
	r.Delete("/users/{id}", DeleteUser)
	return r
}

func Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u, err := GetOne(db, id)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := DeleteByUID(db, id)
}

func Create(w http.ResponseWriter, r *http.Request) {
}

func Update(w http.ResponseWriter, r *http.Request) {
}
