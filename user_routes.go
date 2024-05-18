package usermod

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Router struct {
	db *gorm.DB
}

func NewRouter(db *gorm.DB) *chi.Mux {
	rr := Router{db: db}
	r := chi.NewRouter()
	r.Get("/users", rr.Get)
	r.Post("/users", rr.CreateUser)
	r.Post("/users/change_password", rr.ChangePassword)
	r.Patch("/users", rr.UpdateUser)
	r.Delete("/users", rr.DeleteUser)
	return r
}

func (rr *Router) Get(w http.ResponseWriter, r *http.Request) {

	uid := r.Context().Value(CTX_UID_KEY).(string)

	u, err := GetOne(rr.db, uid)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(u)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (rr *Router) DeleteUser(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(CTX_UID_KEY).(string)
	err := DeleteByUID(rr.db, uid)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rr *Router) CreateUser(w http.ResponseWriter, r *http.Request) {
	u := User{}
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(bytes, &u)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	err = u.Insert(rr.db)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (rr *Router) UpdateUser(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(CTX_UID_KEY).(string)

	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone_number")

	err := Update(rr.db, uid, name, email, phone)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rr *Router) ChangePassword(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(CTX_UID_KEY).(string)

	newPassword := r.FormValue("password")
	oldPassword := r.FormValue("oldPassword")
	u, err := AuthenticateByUID(rr.db, uid, []byte(oldPassword))
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	err = u.ChangePassword(rr.db, []byte(newPassword))
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
