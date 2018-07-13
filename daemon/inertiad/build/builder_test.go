package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/cfg"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder(cfg.Config{}, nil)
	assert.NotNil(t, b)
}
