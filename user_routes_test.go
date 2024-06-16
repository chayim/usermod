package usermod_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/chayim/usermod"
	"github.com/stretchr/testify/assert"
)

var endpoint = "/api/user"

func (s *UserModTestSuite) TestCreateUserRoute() {

	url := s.ts.URL + endpoint

	u := usermod.User{Name: "Chayim",
		Email:    "c@ummmfoo.com",
		Password: []byte("password!!"),
	}

	b, _ := json.Marshal(u)
	reader := bytes.NewReader(b)

	r, _ := http.NewRequest(http.MethodPost, url, reader)
	w, _ := http.DefaultClient.Do(r)

	assert.Equal(s.T(), http.StatusCreated, w.StatusCode)

	found, err := usermod.GetUserByEmail(s.db, u.Email)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), u.Name, found.Name)
	assert.NotEqual(s.T(), u.Password, found.Password)
}

func (s *UserModTestSuite) TestGetUserRoute() {
	u := s.newActivatedUser()
	url := s.ts.URL + endpoint

	r, _ := http.NewRequest(http.MethodGet, url, nil)
	auth := u.Email + ":" + string(testPassword)
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	r.Header.Add("Authorization", basicAuth)
	w, _ := http.DefaultClient.Do(r)
	assert.Equal(s.T(), http.StatusOK, w.StatusCode)

	bytes, _ := io.ReadAll(w.Body)
	assert.True(s.T(), strings.Index(string(bytes), u.Email) > 0)
	assert.True(s.T(), strings.Index(string(bytes), u.Name) > 0)
}

func (s *UserModTestSuite) TestDeleteUserRoute() {
	u := s.newActivatedUser()
	url := s.ts.URL + endpoint

	// is this possible?
	r, _ := http.NewRequest(http.MethodDelete, url, nil)

	auth := u.Email + ":" + string(testPassword)
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	r.Header.Add("Authorization", basicAuth)

	w, _ := http.DefaultClient.Do(r)
	assert.Equal(s.T(), http.StatusOK, w.StatusCode)

	found, err := usermod.GetUserByID(s.db, u.ID.String())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), u.ID, found.ID)
	assert.Equal(s.T(), u.IsDeleted, false)
}

func (s *UserModTestSuite) TestUpdateUserRoute() {
	u := s.newActivatedUser()
	url := s.ts.URL + endpoint

	r, _ := http.NewRequest(http.MethodPatch, url, nil)

	// auth check
	w, _ := http.DefaultClient.Do(r)
	assert.GreaterOrEqual(s.T(), w.StatusCode, 400)

	// is this possible?
	uu := usermod.UpdateJSON{Name: "Bob Dobbs", Phone: "41677791231"}
	b, _ := json.Marshal(uu)
	reader := bytes.NewReader(b)

	auth := u.Email + ":" + string(testPassword)
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	r, _ = http.NewRequest(http.MethodPatch, url, reader)
	r.Header.Add("Authorization", basicAuth)
	w, _ = http.DefaultClient.Do(r)
	assert.Equal(s.T(), http.StatusOK, w.StatusCode)

	found, err := usermod.GetUserByEmail(s.db, u.Email)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), uu.Name, found.Name)
	assert.Equal(s.T(), uu.Phone, found.PhoneNumber)
}

func (s *UserModTestSuite) TestChangePasswordRoute() {
	u := s.newActivatedUser()
	url := s.ts.URL + "/api/change_password"

	// is this possible?
	uu := usermod.PasswordJSON{OldPassword: testPassword, NewPassword: []byte("thisisnewyou")}
	b, _ := json.Marshal(uu)
	reader := bytes.NewReader(b)

	r, _ := http.NewRequest(http.MethodPost, url, reader)
	auth := u.Email + ":" + string(testPassword)
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	r.Header.Add("Authorization", basicAuth)
	w, _ := http.DefaultClient.Do(r)
	assert.Equal(s.T(), http.StatusOK, w.StatusCode)

	u2, err := usermod.AuthenticateByEmail(s.db, u.Email, uu.NewPassword)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), u2.Email, u.Email)
}

func (s *UserModTestSuite) TestActivatingUser() {
	u := s.newUser()
	url := s.ts.URL + "/api/user/activate"
	uot := usermod.NewUserOperationTokenDefaultExpires(s.db,
		u.ID, usermod.ActivationToken)
	err := uot.Insert()
	assert.Nil(s.T(), err)

	tests := []struct {
		name       string
		qstring    string
		statusCode int
	}{{
		name:       "no query string",
		qstring:    "",
		statusCode: http.StatusBadRequest,
	}, {
		name:       "no token",
		qstring:    "email=c@foo.com",
		statusCode: http.StatusBadRequest,
	}, {
		name:       "token only",
		qstring:    "token=123213213",
		statusCode: http.StatusNotFound,
	}, {
		name:       "ok",
		qstring:    fmt.Sprintf("token=%s", uot.ID.String()),
		statusCode: http.StatusOK,
	}}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			uri := url + "?" + tc.qstring
			r, _ := http.NewRequest(http.MethodGet, uri, nil)
			w, _ := http.DefaultClient.Do(r)
			assert.Equal(s.T(), tc.statusCode, w.StatusCode)
		})
	}

}

func (s *UserModTestSuite) TestForgotPassword() {
	u := s.newActivatedUser()
	url := s.ts.URL + "/api/user/forgot_password"
	tests := []struct {
		name       string
		email      string
		statusCode int
	}{{
		name:       "no email",
		email:      "",
		statusCode: http.StatusBadRequest,
	}, {
		name:       "bad email",
		email:      "c@asasdsadsa.com",
		statusCode: http.StatusNotFound,
	}, {
		name:       "works",
		email:      u.Email,
		statusCode: http.StatusOK,
	}}
	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, url+"?email="+tc.email, nil)
			w, _ := http.DefaultClient.Do(r)
			assert.Equal(s.T(), w.StatusCode, tc.statusCode)
		})
	}

}
