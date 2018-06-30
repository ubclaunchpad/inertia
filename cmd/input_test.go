package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/cfg"
)

func TestRemoteAddWalkthrough(t *testing.T) {
	config := cfg.NewConfig("", "", "", "")
	in, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	defer in.Close()

	fmt.Fprintln(in, "pemfile")
	fmt.Fprintln(in, "user")
	fmt.Fprintln(in, "0.0.0.0")
	fmt.Fprintln(in, "master")

	_, err = in.Seek(0, io.SeekStart)
	assert.Nil(t, err)

	err = addRemoteWalkthrough(in, config, "inertia-rocks", "8080", "22", "dev")
	r, found := config.GetRemote("inertia-rocks")
	assert.True(t, found)
	assert.Equal(t, "pemfile", r.PEM)
	assert.Equal(t, "user", r.User)
	assert.Equal(t, "0.0.0.0", r.IP)
	assert.Nil(t, err)
}

func TestRemoteAddWalkthroughFailure(t *testing.T) {
	config := cfg.NewConfig("", "", "", "")
	in, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	defer in.Close()

	fmt.Fprintln(in, "pemfile")
	fmt.Fprintln(in, "")

	_, err = in.Seek(0, io.SeekStart)
	assert.Nil(t, err)

	err = addRemoteWalkthrough(in, config, "inertia-rocks", "8080", "22", "dev")
	assert.Equal(t, errInvalidUser, err)

	in.WriteAt([]byte("pemfile\nuser\n\n"), 0)
	_, err = in.Seek(0, io.SeekStart)
	assert.Nil(t, err)

	err = addRemoteWalkthrough(in, config, "inertia-rocks", "8080", "22", "dev")
	assert.Equal(t, errInvalidAddress, err)
}

func Test_addProjectWalkthrough(t *testing.T) {
	tests := []struct {
		name              string
		wantBuildType     string
		wantBuildFilePath string
		wantErr           bool
	}{
		{"invalid build type", "", "", true},
		{"invalid build file path", "dockerfile", "", true},
		{"herokuish", "herokuish", "", false},
		{"docker-compose", "docker-compose", "docker-compose.yml", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, err := ioutil.TempFile("", "")
			assert.Nil(t, err)
			defer in.Close()

			fmt.Fprintln(in, tt.wantBuildType)
			fmt.Fprintln(in, tt.wantBuildFilePath)

			_, err = in.Seek(0, io.SeekStart)
			assert.Nil(t, err)

			gotBuildType, gotBuildFilePath, err := addProjectWalkthrough(in)
			if (err != nil) != tt.wantErr {
				t.Errorf("addProjectWalkthrough() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotBuildType != tt.wantBuildType {
					t.Errorf("addProjectWalkthrough() gotBuildType = %v, want %v", gotBuildType, tt.wantBuildType)
				}
				if gotBuildFilePath != tt.wantBuildFilePath {
					t.Errorf("addProjectWalkthrough() gotBuildFilePath = %v, want %v", gotBuildFilePath, tt.wantBuildFilePath)
				}
			}
		})
	}
}
