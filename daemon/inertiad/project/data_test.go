package project

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataManager_EnvVariableOperations(t *testing.T) {
	type args struct {
		name    string
		value   string
		encrypt bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no encrypt", args{"myvar1", "mysekret", false}, false},
		{"encrypt", args{"myvar2", "myothersekret", true}, false},
		{"invalid env", args{"", "", true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := "./test_config"
			err := os.Mkdir(dir, os.ModePerm)
			assert.Nil(t, err)
			defer os.RemoveAll(dir)

			c, err := newDataManager(path.Join(dir, "deployment.db"))
			assert.Nil(t, err)

			err = c.AddEnvVariable(tt.args.name, tt.args.value, tt.args.encrypt)
			assert.Equal(t, tt.wantErr, (err != nil))

			vars, err := c.getEnvVariables()
			assert.Nil(t, err)
			assert.Equal(t, tt.args.value, vars[tt.args.name])
		})
	}
}
