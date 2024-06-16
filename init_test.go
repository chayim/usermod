package usermod_test

import (
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/chayim/usermod"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
)

type UserModTestSuite struct {
	ts *httptest.Server
	db *sql.DB
	suite.Suite
}

// func (suite *UserModTestSuite) BeforeTest(suiteName, testName string) {
func (suite *UserModTestSuite) SetupTest() {

	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		log.Fatal(err)
	}
	r := chi.NewRouter()
	suite.db = db

	usermod.CreateAllTables(suite.db)

	r2 := usermod.NewRouter(suite.db)
	r.Mount("/api", r2)

	r.With(usermod.BasicAuth(suite.db)).Get("/auth", testingEndpoint)
	// r.With(usermod.JWTTokenAuth(testingEndpoint)).Get("/auth2", testingEndpoint)
	suite.ts = httptest.NewServer(r)
}

func (suite *UserModTestSuite) AfterTest(suiteName, testName string) {
	suite.ts.Close()
	suite.db.Close()
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(UserModTestSuite))
}

func testingEndpoint(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
