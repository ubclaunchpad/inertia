package containers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDockerClient(t *testing.T) {
	c, err := NewDockerClient()
	assert.Nil(t, err)
	assert.NotNil(t, c)
}
