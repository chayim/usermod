package usermod_test

import (
	"testing"

	"github.com/chayim/usermod"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var testPassword = []byte("iamapasword")

func (s *UserModTestSuite) newUser() *usermod.User {
	u := usermod.NewUser("Chayim", "c@kirshen.com", testPassword)
	err := u.Insert(s.db)
	assert.Nil(s.T(), err)
	return u
}

func (s *UserModTestSuite) TestCreateGetDelete() {
	u := s.newUser()

	dRes := usermod.DeleteByUID(s.db, u.ID.String())
	assert.Nil(s.T(), dRes)
	_, err := usermod.GetOne(s.db, u.ID.String())
	assert.NotNil(s.T(), err)

	u2 := s.newUser()
	u2.Insert(s.db)
	sRes := usermod.SoftDeleteByUID(s.db, u2.ID.String())
	assert.Nil(s.T(), sRes)

	// soft delete check
	s.db.Take(&u2)
	assert.True(s.T(), u2.IsDeleted)
}

func (s *UserModTestSuite) TestActivateUser() {
	u := s.newUser()
	assert.False(s.T(), u.IsActivated)

	err := usermod.ActivateUser(s.db, u.Email)
	assert.Nil(s.T(), err)

	found, _ := usermod.GetOne(s.db, u.ID.String())
	assert.True(s.T(), found.IsActivated)
}

func (s *UserModTestSuite) TestCreateUpdateGet() {
	u := s.newUser()

	err := usermod.Update(s.db, u.ID.String(), "", "not@email.com", "")
	assert.Nil(s.T(), err)

	u2, err := usermod.GetOne(s.db, u.ID.String())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), u2.Email, "not@email.com")
	assert.Equal(s.T(), u2.Name, u.Name)
	assert.Equal(s.T(), u2.PhoneNumber, u.PhoneNumber)
}

func (s *UserModTestSuite) TestChangePassword() {
	u := s.newUser()
	original := u.Password

	err := u.ChangePassword(s.db, []byte("potatofurby"))
	assert.Nil(s.T(), err)

	found, err := usermod.GetOne(s.db, u.ID.String())
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), string(original), string(found.Password))
}

func (s *UserModTestSuite) TestAuthentication() {
	user := s.newUser()

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
