package usermod

import (
	"encoding/json"
	"errors"
	"net/http"
)

func jsonError(w http.ResponseWriter, err error, code int) {
	type f struct {
		Message string `json:"error"`
	}
	e := f{Message: err.Error()}
	bytes, _ := json.Marshal(&e)
	http.Error(w, string(bytes), code)
}

func jsonErrorFromString(w http.ResponseWriter, s string, code int) {
	err := errors.New(s)
	jsonError(w, err, code)
}
