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
	manager, err := newUserManager("./test/test.db")
	assert.Nil(t, err)
	assert.NotNil(t, manager)

	err = manager.AddUser("bobheadxi", "best_person_ever")
	assert.Nil(t, err)

	correct, err := manager.IsCorrectCredentials("bobheadxi", "not_quite_best")
	assert.Nil(t, err)
	assert.False(t, correct)

	correct, err = manager.IsCorrectCredentials("bobheadxi", "best_person_ever")
	assert.Nil(t, err)
	assert.True(t, correct)
}
