package input

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
			gotBuildType, gotBuildFilePath, err := AddProjectWalkthrough()
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

func Test_enterEC2CredentialsWalkthrough(t *testing.T) {
	tests := []struct {
		name    string
		wantID  string
		wantKey string
		wantErr bool
	}{
		{"bad ID", "", "asdf", true},
		{"bad key", "asdf", "", true},
		{"good", "asdf", "asdf", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, err := ioutil.TempFile("", "")
			assert.NoError(t, err)
			defer in.Close()

			fmt.Fprintln(in, tt.wantID)
			fmt.Fprintln(in, tt.wantKey)

			_, err = in.Seek(0, io.SeekStart)
			assert.NoError(t, err)

			var old = os.Stdin
			os.Stdin = in
			defer func() { os.Stdin = old }()
			gotID, gotKey, err := EnterEC2CredentialsWalkthrough()
			if (err != nil) != tt.wantErr {
				t.Errorf("enterEC2CredentialsWalkthrough() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotID != tt.wantID {
					t.Errorf("enterEC2CredentialsWalkthrough() gotId = %v, want %v", gotID, tt.wantID)
				}
				if gotKey != tt.wantKey {
					t.Errorf("enterEC2CredentialsWalkthrough() gotKey = %v, want %v", gotKey, tt.wantKey)
				}
			}
		})
	}
}

func Test_chooseFromListWalkthrough(t *testing.T) {
	type args struct {
		optionName string
		options    []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"bad", args{"chickens", []string{"deep fried", "baked"}}, "", true},
		{"good", args{"chickens", []string{"deep fried", "baked"}}, "baked", false},
	}
	for _, tt := range tests {
		in, err := ioutil.TempFile("", "")
		assert.NoError(t, err)
		defer in.Close()

		fmt.Fprintln(in, tt.want)

		_, err = in.Seek(0, io.SeekStart)
		assert.NoError(t, err)

		var old = os.Stdin
		os.Stdin = in
		defer func() { os.Stdin = old }()
		t.Run(tt.name, func(t *testing.T) {
			got, err := ChooseFromListWalkthrough(tt.args.optionName, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("chooseFromListWalkthrough() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got != tt.want {
					t.Errorf("chooseFromListWalkthrough() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
