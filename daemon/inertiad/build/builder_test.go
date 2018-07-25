package build

import (
	"context"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/cfg"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder(cfg.Config{}, nil)
	assert.NotNil(t, b)
}

const (
	DockerComposeVersion = "docker/compose:1.21.0"
	HerokuishVersion     = "gliderlabs/herokuish:v0.4.0"
)

// killTestContainers is a helper for tests - it implements project.ContainerStopper
func killTestContainers(cli *docker.Client, w io.Writer) error {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	// Take down all containers except the testvps
	for _, container := range containers {
		if container.Names[0] != "/testvps" {
			err := cli.ContainerKill(ctx, container.ID, "SIGKILL")
			if err != nil {
				return err
			}
		}
	}

	// Prune images
	_, err = cli.ContainersPrune(ctx, filters.Args{})
	return err
}

func TestBuilder_Build(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		buildType string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"type docker-compose", args{"docker-compose"}, false},
		{"type dockerfile", args{"dockerfile"}, false},
		{"type herokuish", args{"herokuish"}, false},
	}

	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				testProjectName = "test_" + tt.args.buildType
				testProjectDir  = path.Join(
					os.Getenv("GOPATH"),
					"/src/github.com/ubclaunchpad/inertia/test/build/"+tt.args.buildType,
				)
				b = NewBuilder(cfg.Config{
					ProjectDirectory:     testProjectDir,
					DockerComposeVersion: DockerComposeVersion,
					HerokuishVersion:     HerokuishVersion,
				}, killTestContainers)
				out = os.Stdout
			)
			deploy, err := b.Build(tt.args.buildType, Config{
				Name:           testProjectName,
				BuildDirectory: testProjectDir,
			}, cli, out)
			if err != nil {
				if tt.wantErr {
					t.Errorf("Builder.Build() error = %v, wantErr %v", err, tt.wantErr)
				} else {
					t.Errorf("unexpected error %v", err)
				}
				return
			}

			// Run containers and watch for abrupt failure
			endTest := false
			errCh := deploy()
			go func() {
				select {
				case err := <-errCh:
					if err != nil && !endTest {
						t.Errorf("unexpected error %v", err)
						return
					}
				}
			}()

			// Arbitrary wait for containers to start
			time.Sleep(10 * time.Second)

			// Check for containers
			containers, err := cli.ContainerList(
				context.Background(),
				types.ContainerListOptions{},
			)
			assert.Nil(t, err)
			foundP := false
			for _, c := range containers {
				if strings.Contains(c.Names[0], testProjectName) {
					foundP = true
				}
			}

			// try again if project no up (workaround for Travis)
			if !foundP {
				time.Sleep(30 * time.Second)
				containers, err = cli.ContainerList(
					context.Background(),
					types.ContainerListOptions{},
				)
				assert.Nil(t, err)
				for _, c := range containers {
					if strings.Contains(c.Names[0], testProjectName) {
						foundP = true
					}
				}
			}

			// try again if project no up (another workaround for Travis)
			if !foundP {
				time.Sleep(180 * time.Second)
				containers, err = cli.ContainerList(
					context.Background(),
					types.ContainerListOptions{},
				)
				assert.Nil(t, err)
				for _, c := range containers {
					if strings.Contains(c.Names[0], testProjectName) {
						foundP = true
					}
				}
			}

			assert.True(t, foundP, "project container should be active")

			endTest = true
			err = killTestContainers(cli, nil)
			assert.Nil(t, err)
		})
	}
}
