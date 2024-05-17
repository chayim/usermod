package usermod

import (
	"encoding/json"
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
