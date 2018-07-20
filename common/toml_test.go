package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadProjectConfig(t *testing.T) {
	version := "latest"
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    InertiaProject
		wantErr bool
	}{
		{"no file found", args{""}, InertiaProject{}, true},
		{"invalid file", args{"inertia.go"}, InertiaProject{}, true},
		{"success", args{"example.inertia.toml"}, InertiaProject{
			Version: &version,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set path relative to root of project
			tt.args.filepath = filepath.Join(filepath.Dir(cwd), tt.args.filepath)

			got, err := ReadProjectConfig(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadProjectConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.NotNil(t, got.Version)
				assert.Equal(t, *got.Version, *tt.want.Version)
			}
		})
	}
}
