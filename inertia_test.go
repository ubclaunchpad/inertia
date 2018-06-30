package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/cmd"
)

func TestMain(t *testing.T) {
	main()
	assert.Equal(t, "latest", cmd.Root.Version)
}

func TestMainSetVersion(t *testing.T) {
	Version = "robert"
	main()
	assert.Equal(t, Version, cmd.Root.Version)
	Version = ""
}
