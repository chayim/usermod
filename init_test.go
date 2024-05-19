package usermod_test

import (
	"net/http/httptest"
	"testing"

	"github.com/chayim/usermod"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type UserModTestSuite struct {
	ts *httptest.Server
	db *gorm.DB
	suite.Suite
}

// func (suite *UserModTestSuite) BeforeTest(suiteName, testName string) {
func (suite *UserModTestSuite) SetupTest() {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Error)})
	r := chi.NewRouter()
	suite.db = db

	// migrate tables
	db.AutoMigrate(&usermod.User{})

	r2 := usermod.NewRouter(suite.db)
	r.Mount("/api", r2)
	suite.ts = httptest.NewServer(r)

}

func (suite *UserModTestSuite) AfterTest(suiteName, testName string) {
	suite.ts.Close()
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(UserModTestSuite))
}
