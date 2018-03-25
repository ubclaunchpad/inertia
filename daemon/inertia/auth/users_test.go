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
	return newUserManager(path.Join(dir, "users.db"), "127.0.0.1", "/", 120)
}

func TestAddUserAndIsCorrectCredentials(t *testing.T) {
	dir := "./test_users"
	manager, err := getTestUserManager(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer manager.Close()

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

	correct, err := manager.IsCorrectCredentials("bobheadxi", "not_quite_best")
	assert.Nil(t, err)
	assert.False(t, correct)

	correct, err = manager.IsCorrectCredentials("bobheadxi", "best_person_ever")
	assert.Nil(t, err)
	assert.True(t, correct)
}

func TestAddDeleteAndHasUser(t *testing.T) {
	dir := "./test_users"
	manager, err := getTestUserManager(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer manager.Close()

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

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
