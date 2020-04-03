package build

import (
	"context"
	"io"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/cfg"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder(cfg.Config{}, nil)
	assert.NotNil(t, b)
}

// killTestContainers is a helper for tests - it implements project.ContainerStopper
func killTestContainers(cli *docker.Client, w io.Writer) error {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	// Take down all containers except the testvps and testcontainer
	for _, container := range containers {
		if container.Names[0] != "/testvps" && container.Names[0] != "/testcontainer" {
			err := cli.ContainerKill(ctx, container.ID, "SIGKILL")
			if err != nil {
				return err
			}
		}
	}

	return err
}

func TestBuilder_Build(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		buildType     string
		buildFilePath string
		persistPath   string
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		expectedErrMsg string
	}{
		{"type docker-compose", args{"docker-compose", "", ""}, false, ""},

		{"type dockerfile", args{"dockerfile", "", ""}, false, ""},
		{"type dockerfile should fail", args{"dockerfile", "fail.Dockerfile", ""}, true, "image build failed"},

		{"type dockerfile with persist", args{"dockerfile", "", "persist"}, false, ""},
		{"type docker-compose with persist", args{"dockerfile", "", "persist"}, false, ""},
	}

	// Setup
	cli, err := containers.NewDockerClient()
	require.NoError(t, err)
	defer cli.Close()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	// Run cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			cli.ContainersPrune(context.Background(), filters.Args{})
			time.Sleep(5 * time.Second)

			var (
				// Generate random project name
				testProjectName = "test_" + tt.args.buildType + "_" + strconv.Itoa(rand.Intn(100))
				// Docker mounts require an absolute path
				testProjectDir = path.Clean(path.Join(cwd, "../../../test/build/"+tt.args.buildType))

				b = NewBuilder(cfg.Config{
					DockerComposeVersion: "docker/compose:latest",
				}, killTestContainers)
				out = os.Stdout
			)

			// set up test persist dir
			if tt.args.persistPath != "" {
				tt.args.persistPath = path.Clean(path.Join(cwd, "../../../test/build/tmp"+tt.args.persistPath))
			}

			// Run build
			t.Logf("Preparing to build test project with name '%s' from directory '%s'",
				testProjectName, testProjectDir)
			deploy, err := b.Build(tt.args.buildType, Config{
				Name:             testProjectName,
				BuildFilePath:    tt.args.buildFilePath,
				BuildDirectory:   testProjectDir,
				PersistDirectory: tt.args.persistPath,
			}, cli, out)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				return
			}
			assert.NoError(t, err)

			// Run containers
			err = deploy()
			assert.NoError(t, err)

			// Arbitrary wait for containers to start
			time.Sleep(10 * time.Second)

			// Check for containers
			containers, err := cli.ContainerList(
				context.Background(),
				types.ContainerListOptions{},
			)
			assert.NoError(t, err)
			foundP := false
			for _, c := range containers {
				if strings.Contains(c.Names[0], testProjectName) {
					foundP = true
				}
			}

			// Wait for project to come up
			attempts := 0
			for !foundP && attempts < 10 {
				attempts++
				time.Sleep(30 * time.Second)
				containers, err = cli.ContainerList(
					context.Background(),
					types.ContainerListOptions{},
				)
				assert.NoError(t, err)
				for _, c := range containers {
					if strings.Contains(c.Names[0], testProjectName) {
						foundP = true
					}
				}
			}
			assert.True(t, foundP, "project container should be active")

			// clean up
			err = killTestContainers(cli, nil)
			assert.NoError(t, err)
			cli.ContainersPrune(context.Background(), filters.Args{})
			time.Sleep(5 * time.Second)
		})
	}
}
