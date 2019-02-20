package input

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
	assert.NoError(t, err)
	defer in.Close()

	fmt.Fprintln(in, "pemfile")
	fmt.Fprintln(in, "user")
	fmt.Fprintln(in, "0.0.0.0")
	fmt.Fprintln(in, "master")

	_, err = in.Seek(0, io.SeekStart)
	assert.NoError(t, err)

	err = AddRemoteWalkthrough(in, config, "inertia-rocks", "8080", "22", "dev")
	r, found := config.GetRemote("inertia-rocks")
	assert.True(t, found)
	assert.Equal(t, "pemfile", r.PEM)
	assert.Equal(t, "user", r.User)
	assert.Equal(t, "0.0.0.0", r.IP)
	assert.NoError(t, err)
}

func TestRemoteAddWalkthroughFailure(t *testing.T) {
	config := cfg.NewConfig("", "", "", "")
	in, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer in.Close()

	fmt.Fprintln(in, "pemfile")
	fmt.Fprintln(in, "")

	_, err = in.Seek(0, io.SeekStart)
	assert.NoError(t, err)

	err = AddRemoteWalkthrough(in, config, "inertia-rocks", "8080", "22", "dev")
	assert.Equal(t, errInvalidUser, err)

	in.WriteAt([]byte("pemfile\nuser\n\n"), 0)
	_, err = in.Seek(0, io.SeekStart)
	assert.NoError(t, err)

	err = AddRemoteWalkthrough(in, config, "inertia-rocks", "8080", "22", "dev")
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

			gotBuildType, gotBuildFilePath, err := AddProjectWalkthrough(in)
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

			gotID, gotKey, err := EnterEC2CredentialsWalkthrough(in)
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

		t.Run(tt.name, func(t *testing.T) {
			got, err := ChooseFromListWalkthrough(in, tt.args.optionName, tt.args.options)
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
