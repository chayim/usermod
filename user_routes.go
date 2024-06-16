package usermod

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	db *sql.DB
}

// NewRouter should be mounted to the correct location within your application
func NewRouter(db *sql.DB) *chi.Mux {
	rr := Router{db: db}
	r := chi.NewRouter()
	r.Post("/user", rr.CreateUser)
	r.With(BasicAuth(db)).Get("/user", rr.Get)
	r.With(BasicAuth(db)).Patch("/user", rr.UpdateUser)
	r.With(BasicAuth(db)).Delete("/user", rr.DeleteUser)
	r.With(BasicAuth(db)).Post("/change_password", rr.ChangePassword)
	r.Get("/user/activate", rr.ActivateUser)
	r.Post("/user/forgot_password", rr.ForgotPassword)

	return r
}

func (rr *Router) Get(w http.ResponseWriter, r *http.Request) {

	uid := r.Context().Value(CTX_USER_KEY).(*User)

	b, err := json.Marshal(uid)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// DelteUser will mark a user as soft deleted in the database
func (rr *Router) DeleteUser(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(CTX_UID_KEY).(string)
	u := User{db: rr.db}
	err := u.SoftDeleteByUID(uid)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rr *Router) CreateUser(w http.ResponseWriter, r *http.Request) {
	u := User{db: rr.db}

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

	err = u.Insert()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	uot := NewUserOperationTokenDefaultExpires(rr.db, u.ID, ActivationToken)
	err = uot.Insert()
	if err != nil {
		jsonError(w, err, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type UpdateJSON struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone_number"`
}

func (rr *Router) UpdateUser(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(CTX_USER_KEY).(*User)

	uu := UpdateJSON{}
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(bytes, &uu)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	name := u.Name
	email := u.Email
	phone := u.PhoneNumber
	if uu.Name != "" {
		name = uu.Name
	}
	if uu.Email != "" {
		email = uu.Email
	}
	if uu.Phone != "" {
		phone = uu.Phone
	}
	err = u.Update(name, email, phone)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type PasswordJSON struct {
	OldPassword []byte `json:"password"`
	NewPassword []byte `json:"newPassword"`
}

func (rr *Router) ChangePassword(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(CTX_USER_KEY).(*User)

	uu := PasswordJSON{}
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(bytes, &uu)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	err = u.ChangePassword(uu.NewPassword)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rr *Router) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		jsonErrorFromString(w, "no email specified", http.StatusBadRequest)
		return
	}
	u, err := GetActiveUserByEmail(rr.db, email)
	if err != nil {
		jsonError(w, err, http.StatusNotFound)
		return
	}

	uot := NewUserOperationTokenDefaultExpires(rr.db, u.ID, ForgotPaswordToken)
	err = uot.Insert()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rr *Router) ActivateUser(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if token == "" {
		jsonErrorFromString(w, "Token must be specified", http.StatusBadRequest)
		return
	}

	uid, err := MarkTokenAsUsed(rr.db, token)
	if err != nil {
		jsonErrorFromString(w, "Invalid token", http.StatusNotFound)
		return
	}

	err = Activate(rr.db, uid.String())
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
