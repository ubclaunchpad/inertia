package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/cfg"
)

func Test_addProjectWalkthrough(t *testing.T) {
	tests := []struct {
		name              string
		wantBuildType     cfg.BuildType
		wantBuildFilePath string
		wantErr           bool
	}{
		{"invalid build type", "", "", true},
		{"invalid build file path", "dockerfile", "", true},
		{"docker-compose", "docker-compose", "docker-compose.yml", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, err := ioutil.TempFile("", "")
			assert.NoError(t, err)
			defer in.Close()

			fmt.Fprintln(in, tt.wantBuildType)
			fmt.Fprintln(in, tt.wantBuildFilePath)

			_, err = in.Seek(0, io.SeekStart)
			assert.NoError(t, err)

			var old = os.Stdin
			os.Stdin = in
			defer func() { os.Stdin = old }()
			gotBuildType, gotBuildFilePath, err := addProjectWalkthrough()
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
