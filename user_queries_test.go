package usermod_test

import (
	"testing"

	"github.com/chayim/usermod"
	"github.com/stretchr/testify/assert"
)

var testPassword = []byte("iamapasword")

func (s *UserModTestSuite) newUser() *usermod.User {
	u := usermod.NewUserWithDetails(s.db, "Chayim", "c@ummmfoo.com", testPassword)
	err := u.Insert()
	assert.Nil(s.T(), err)
	return u
}

func (s *UserModTestSuite) newActivatedUser() *usermod.User {
	u := s.newUser()
	usermod.Activate(s.db, u.ID.String())
	return u
}

func (s *UserModTestSuite) TestUserCreateGetDelete() {
	u := s.newUser()
	dRes := u.DeleteByUID(u.ID.String())
	assert.Nil(s.T(), dRes)

	_, err := usermod.GetUserByID(s.db, u.ID.String())
	assert.NotNil(s.T(), err)

	u2 := s.newUser()
	err = u2.SoftDeleteByUID(u2.ID.String())
	assert.Nil(s.T(), err)
	u3, err := usermod.GetUserByID(s.db, u2.ID.String())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), u3.ID, u2.ID)
	assert.Equal(s.T(), u3.Name, u2.Name)
	assert.Equal(s.T(), u3.Email, u2.Email)
	assert.Equal(s.T(), u3.IsDeleted, true)

	// now get by email
	u4, err := usermod.GetUserByEmail(s.db, u2.Email)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), u4.ID, u2.ID)
	assert.Equal(s.T(), u4.Name, u2.Name)
	assert.Equal(s.T(), u4.Email, u2.Email)

}

func (s *UserModTestSuite) TestActivateUser() {
	u := s.newUser()
	assert.False(s.T(), u.IsActivated)

	err := usermod.Activate(s.db, u.ID.String())
	assert.Nil(s.T(), err)

	found, err := usermod.GetUserByID(s.db, u.ID.String())
	assert.Nil(s.T(), err)
	assert.True(s.T(), found.IsActivated)
}

func (s *UserModTestSuite) TestUserCreateUpdateGet() {
	u := s.newUser()

	err := u.Update("", "not@email.com", "")
	assert.Nil(s.T(), err)

	u2, err := usermod.GetUserByID(s.db, u.ID.String())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), u2.Email, "not@email.com")
	assert.Equal(s.T(), u2.Name, u.Name)
	assert.Equal(s.T(), u2.PhoneNumber, u.PhoneNumber)
}

func (s *UserModTestSuite) TestUserChangePassword() {
	u := s.newUser()
	original := u.Password

	err := u.ChangePassword([]byte("potatofurby"))
	assert.Nil(s.T(), err)

	found, err := usermod.GetUserByID(s.db, u.ID.String())
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), string(original), string(found.Password))
}

func (s *UserModTestSuite) TestUserAuthenticationByUID() {
	user := s.newActivatedUser()

	tests := []struct {
		name     string
		val      string
		password []byte
		expected bool
	}{{}, {
		name:     "ID auth should work",
		val:      user.ID.String(),
		password: testPassword,
		expected: true,
	}, {
		name:     "Right ID, wrong password",
		val:      user.ID.String(),
		password: []byte("notthepassword"),
		expected: false,
	}, {
		name:     "Invalid ID",
		val:      "just-a-bad-id",
		password: testPassword,
		expected: false,
	}}
	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			u, err := usermod.AuthenticateByUID(s.db, tc.val, tc.password)
			if tc.expected {
				assert.Nil(t, err)
				assert.Equal(t, user.Email, u.Email)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, "", u.Email)
			}
		})
	}

}

func (s *UserModTestSuite) TestUserAuthenticationByEmail() {
	user := s.newActivatedUser()

	tests := []struct {
		name     string
		val      string
		password []byte
		expected bool
	}{{
		name:     "Email auth should work",
		val:      user.Email,
		password: testPassword,
		expected: true,
	}, {
		name:     "Right email, bad password",
		val:      user.Email,
		password: []byte("notthepassword"),
		expected: false,
	}, {
		name:     "Invalid email",
		val:      "iam@invalid.com",
		password: testPassword,
		expected: false,
	}}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			u, err := usermod.AuthenticateByEmail(s.db, tc.val, tc.password)
			if tc.expected {
				assert.Nil(t, err)
				assert.Equal(t, user.Email, u.Email)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, "", u.Email)
			}
		})
	}
}
