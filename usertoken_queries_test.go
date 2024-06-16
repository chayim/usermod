package usermod_test

import (
	"github.com/chayim/usermod"
	"github.com/stretchr/testify/assert"
)

func (s *UserModTestSuite) newUserOperationsToken(user *usermod.User) *usermod.UserOperationToken {
	u := usermod.NewUserOperationTokenDefaultExpires(s.db, user.ID, usermod.ForgotPaswordToken)
	err := u.Insert()
	assert.Nil(s.T(), err)
	return u
}

func (s *UserModTestSuite) TestUOTCreateGet() {
	user := s.newUser()
	u := s.newUserOperationsToken(user)

	res, err := usermod.GetUserOperationToken(s.db, u.ID.String())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), u.ID, res.ID)

}

func (s *UserModTestSuite) TestUOTGetTokenIfValid() {
	user := s.newUser()
	u := s.newUserOperationsToken(user)
	assert.Equal(s.T(), u.Used, false)

	res, err := usermod.GetTokenIfValid(s.db, user.ID.String(), u.ID.String())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), res.Used, false)
}

func (s *UserModTestSuite) TestUOTMarkTokenAsUsed() {
	user := s.newUser()
	u := s.newUserOperationsToken(user)

	uid, err := usermod.MarkTokenAsUsed(s.db, u.ID.String())
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), uid)

	res, _ := usermod.GetUserOperationToken(s.db, u.ID.String())
	assert.Equal(s.T(), res.Used, true)
}
