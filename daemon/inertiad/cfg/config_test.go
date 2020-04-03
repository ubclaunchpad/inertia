package cfg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	os.Setenv("INERTIA_PROJECT_DIR", "/user/project")
	cfg := New()
	assert.Equal(t, "/user/project", cfg.ProjectDirectory)
	t.Log(cfg.DockerComposeVersion)
}
