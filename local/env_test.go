package local

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHomePath(t *testing.T) {
	env, err := GetHomePath()
	assert.NoError(t, err)
	assert.NotEqual(t, env, "")
	assert.DirExists(t, env)
}
