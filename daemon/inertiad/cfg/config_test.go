package cfg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	os.Setenv("INERTIA_SSL_DIR", "/user/ssl")
	cfg := New()
	assert.Equal(t, "/user/ssl", cfg.SSLDirectory)
}
