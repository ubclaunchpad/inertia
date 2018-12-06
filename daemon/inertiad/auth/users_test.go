package auth

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestUserManager(dir string) (*userManager, error) {
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return newUserManager(path.Join(dir, "users.db"))
}

func TestAddUserAndIsCorrectCredentials(t *testing.T) {
	dir := "./test_users"
	manager, err := getTestUserManager(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer manager.Close()

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

	_, correct, err := manager.IsCorrectCredentials("bobheadxi", "not_quite_best")
	assert.Nil(t, err)
	assert.False(t, correct)

	_, correct, err = manager.IsCorrectCredentials("bobheadxi", "best_person_ever")
	assert.Nil(t, err)
	assert.True(t, correct)
}

func TestAllUserManagementOperations(t *testing.T) {
	dir := "./test_users"
	manager, err := getTestUserManager(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer manager.Close()

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

	err = manager.AddUser("whoisthat", "ummmmmmmmmm", false)
	assert.Nil(t, err)

	users := manager.UserList()
	assert.Equal(t, len(users), 3) // There is a master user in here too

	err = manager.HasUser("bobheadxi")
	assert.Nil(t, err)

	err = manager.RemoveUser("bobheadxi")
	assert.Nil(t, err)

	err = manager.HasUser("bobheadxi")
	assert.Equal(t, errUserNotFound, err)
}

func TestIsAdmin(t *testing.T) {
	dir := "./test_users"
	manager, err := getTestUserManager(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer manager.Close()

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

	admin, err := manager.IsAdmin("bobheadxi")
	assert.Nil(t, err)
	assert.True(t, admin)

	err = manager.AddUser("chadlagore", "chadlad", false)
	assert.Nil(t, err)

	admin, err = manager.IsAdmin("chadlagore")
	assert.Nil(t, err)
	assert.False(t, admin)
}

func TestRemoveUser(t *testing.T) {
	dir := "./test_users"
	manager, err := getTestUserManager(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer manager.Close()

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

	err = manager.RemoveUser("bobheadxi")
	assert.Nil(t, err)

	err = manager.HasUser("bobheadxi")
	assert.NotNil(t, err)
	assert.Equal(t, errUserNotFound, err)
}

func TestTooManyLogins(t *testing.T) {
	dir := "./test_users_login_limit"
	manager, err := getTestUserManager(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer manager.Close()

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

	for i := 0; i < loginAttemptsLimit; i++ {
		_, correct, err := manager.IsCorrectCredentials("bobheadxi", "not_quite_best")
		assert.Nil(t, err)
		assert.False(t, correct)
	}

	_, correct, err := manager.IsCorrectCredentials("bobheadxi", "not_quite_best")
	assert.False(t, correct)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "login attempts")

	err = manager.HasUser("bobheadxi")
	assert.NotNil(t, err)
	assert.Equal(t, errUserNotFound, err)
}

func TestEnableTotp(t *testing.T) {
	dir := "./test_users"
	manager, err := getTestUserManager(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer manager.Close()

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

	manager.EnableTotp("bobheadxi")
	result, err := manager.IsTotpEnabled("bobheadxi")
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestDisableTotp(t *testing.T) {
	dir := "./test_users"
	manager, err := getTestUserManager(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer manager.Close()

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

	manager.EnableTotp("bobheadxi")
	result, err := manager.IsTotpEnabled("bobheadxi")
	assert.Nil(t, err)
	assert.True(t, result)

	manager.DisableTotp("bobheadxi")
	result, err = manager.IsTotpEnabled("bobheadxi")
	assert.Nil(t, err)
	assert.False(t, result)
}

func TestRemoveBackupCode(t *testing.T) {
	dir := "./test_users"
	manager, err := getTestUserManager(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer manager.Close()

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

	// good code
	_, backupCodes, err := manager.EnableTotp("bobheadxi")
	result, err := manager.IsValidBackupCode("bobheadxi", backupCodes[0])
	assert.Nil(t, err)
	assert.True(t, result)

	// bad code
	result, err = manager.IsValidBackupCode("bobheadxi", "abcde-fghij")
	assert.Nil(t, err)
	assert.False(t, result)

	// consume the good code
	err = manager.RemoveBackupCode("bobheadxi", backupCodes[0])
	assert.Nil(t, err)

	// good code should now fail
	result, err = manager.IsValidBackupCode("bobheadxi", backupCodes[0])
	assert.Nil(t, err)
	assert.False(t, result)

	// removing already removed should fail
	err = manager.RemoveBackupCode("bobheadxi", backupCodes[0])
	assert.NotNil(t, err)
}
