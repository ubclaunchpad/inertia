package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionsHandlerConstructor(t *testing.T) {
	dir := "./test"
	err := os.Mkdir(dir, os.ModePerm)
	assert.Nil(t, err)
	defer os.RemoveAll(dir)
	ph, err := NewPermissionsHandler("./test/users.db", nil)
	assert.Nil(t, err)
	assert.NotNil(t, ph)
}
