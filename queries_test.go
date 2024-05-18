package usermod_test

import (
	"testing"

	"github.com/chayim/usermod"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var testPassword = []byte("iamapasword")

func newUser() *usermod.User {
	return usermod.NewUser("Chayim", "c@kirshen.com", testPassword)
}

func (s *UserModTestSuite) TestCreateGetDelete() {
	u := newUser()
	err := u.Insert(s.db)
	assert.Nil(s.T(), err)

	u2 := usermod.User{}
	res := s.db.Take(&u2)
	assert.Nil(s.T(), res.Error)
	assert.Equal(s.T(), u.Email, u2.Email)

	dRes := usermod.DeleteByUID(s.db, u.ID.String())
	assert.Nil(s.T(), dRes)

	u = newUser()
	u.Insert(s.db)
	sRes := usermod.SoftDeleteByUID(s.db, u.ID.String())
	assert.Nil(s.T(), sRes)

	// // soft delete check
	u2 = usermod.User{}
	s.db.Take(&u2)
	assert.True(s.T(), u2.IsDeleted)
}

func (s *UserModTestSuite) TestCreateUpdateGet() {
	u := newUser()
	err := u.Insert(s.db)
	assert.Nil(s.T(), err)

	err = usermod.Update(s.db, u.ID.String(), "", "not@email.com", "")
	assert.Nil(s.T(), err)

	u2 := usermod.User{}
	s.db.Take(&u2)
	assert.Equal(s.T(), u2.Email, "not@email.com")
	assert.Equal(s.T(), u2.Name, u.Name)
	assert.Equal(s.T(), u2.PhoneNumber, u.PhoneNumber)
}

func (s *UserModTestSuite) TestChangePassword() {
	u := newUser()
	err := u.Insert(s.db)
	assert.Nil(s.T(), err)
	original := u.Password

	err = u.ChangePassword(s.db, []byte("potatofurby"))
	assert.Nil(s.T(), err)

	found, err := usermod.GetOne(s.db, u.ID.String())
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), string(original), string(found.Password))
}

func (s *UserModTestSuite) TestAuthentication() {
	user := newUser()
	err := user.Insert(s.db)
	assert.Nil(s.T(), err)

	tests := []struct {
		name     string
		val      string
		password []byte
		expected bool
		f        func(db *gorm.DB, x string, password []byte) (*usermod.User, error)
	}{{
		name:     "Email auth should work",
		val:      user.Email,
		password: testPassword,
		expected: true,
		f:        usermod.AuthenticateByEmail,
	}, {
		name:     "Right email, bad password",
		val:      user.Email,
		password: []byte("notthepassword"),
		expected: false,
		f:        usermod.AuthenticateByEmail,
	}, {
		name:     "Invalid email",
		val:      "iam@invalid.com",
		password: testPassword,
		expected: false,
		f:        usermod.AuthenticateByEmail,
	}, {
		name:     "ID auth should work",
		val:      user.ID.String(),
		password: testPassword,
		expected: true,
		f:        usermod.AuthenticateByUID,
	}, {
		name:     "Right ID, wrong password",
		val:      user.ID.String(),
		password: []byte("notthepassword"),
		expected: false,
		f:        usermod.AuthenticateByUID,
	}, {
		name:     "Invalid ID",
		val:      "just-a-bad-id",
		password: testPassword,
		expected: false,
		f:        usermod.AuthenticateByUID,
	}}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			u, err := tc.f(s.db, tc.val, tc.password)
			if tc.expected {
				assert.Nil(t, err)
				assert.NotNil(t, u)
			} else {
				assert.NotNil(t, err)
				assert.Nil(t, u)
			}
		})
	}

}
