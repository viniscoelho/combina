package lottostore

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// const username = "johndoe"

// func TestUserCreationFailed_EmptyName(t *testing.T) {
// 	assert := assert.New(t)

// 	_, err := NewUser("", "password", "member")
// 	assert.Error(err, "user should not be created: empty name")
// }

// func TestUserCreationFailed_InvalidPassword(t *testing.T) {
// 	assert := assert.New(t)

// 	_, err := NewUser(username, "passwor", "member")
// 	assert.Error(err, "user should not be created: invalid password")
// }

// func TestUserCreationFailed_InvalidRole(t *testing.T) {
// 	assert := assert.New(t)

// 	_, err := NewUser(username, "password", "boss")
// 	assert.Error(err, "user should not be created: invalid role")
// }

// func TestUserCreationSuccess(t *testing.T) {
// 	assert := assert.New(t)

// 	newUser, err := NewUser(username, "password", "member")
// 	assert.NoError(err, "user should be created")
// 	assert.NotNil(newUser)
// }

// func TestPasswordChangeFailed(t *testing.T) {
// 	assert := assert.New(t)

// 	newUser, err := NewUser(username, "password", "member")
// 	assert.NoError(err, "user should be created")
// 	assert.NotNil(newUser)

// 	err = newUser.ChangePassword("shortpw")
// 	assert.Error(err, "new password should not meet the requirements")
// }

// func TestPasswordChangeSuccess(t *testing.T) {
// 	assert := assert.New(t)

// 	newUser, err := NewUser(username, "myPassword", "member")
// 	assert.NoError(err, "user should be created")
// 	assert.NotNil(newUser)

// 	err = newUser.ChangePassword("newPassword")
// 	assert.NoError(err, "new password should meet the requirements")
// }
