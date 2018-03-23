package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddUserAndIsCorrectCredentials(t *testing.T) {
	dir := "./test"
	err := os.Mkdir(dir, os.ModePerm)
	assert.Nil(t, err)
	defer os.RemoveAll(dir)
	manager, err := newUserManager("./test/test.db", 120)
	assert.Nil(t, err)
	assert.NotNil(t, manager)

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
	dir := "./test"
	err := os.Mkdir(dir, os.ModePerm)
	assert.Nil(t, err)
	defer os.RemoveAll(dir)
	manager, err := newUserManager("./test/test.db", 120)
	assert.Nil(t, err)
	assert.NotNil(t, manager)

	err = manager.AddUser("bobheadxi", "best_person_ever", true)
	assert.Nil(t, err)

	found, err := manager.HasUser("bobheadxi")
	assert.Nil(t, err)
	assert.True(t, found)

	err = manager.RemoveUser("bobheadxi")
	assert.Nil(t, err)

	found, err = manager.HasUser("bobheadxi")
	assert.Nil(t, err)
	assert.False(t, found)
}

func TestIsAdmin(t *testing.T) {
	dir := "./test"
	err := os.Mkdir(dir, os.ModePerm)
	assert.Nil(t, err)
	defer os.RemoveAll(dir)
	manager, err := newUserManager("./test/test.db", 120)
	assert.Nil(t, err)
	assert.NotNil(t, manager)

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
