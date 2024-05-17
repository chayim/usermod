package usermod_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chayim/usermod"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestSuite struct {
	ts *httptest.Server
	db *gorm.DB
	suite.Suite
}

func (suite *TestSuite) BeforeTest(suiteName, testName string) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	r := chi.NewRouter()
	suite.db = db

	// migrate tables
	db.AutoMigrate(&usermod.User{}, &usermod.UserAddress{})

	r2 := usermod.NewRouter(suite.db)
	r.Mount("/api", r2)
	suite.ts = httptest.NewServer(r)

}

func (suite *TestSuite) AfterTest(suiteName, testName string) {
	suite.ts.Close()
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
