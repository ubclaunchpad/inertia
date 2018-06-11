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
		decrypt bool
		wantErr bool
	}{
		{"no encrypt", args{"myvar1", "mysekret", false}, true, false},
		{"encrypt", args{"myvar2", "myothersekret", true}, true, false},
		{"invalid env", args{"", "", true}, true, true},
		{"no decrypt", args{"myvar", "asdfasdf", true}, true, false},
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

			vars, err := c.GetEnvVariables(tt.decrypt)
			assert.Nil(t, err)

			if tt.decrypt {
				assert.Equal(t, tt.args.value, vars[tt.args.name])
			} else {
				assert.Equal(t, "[ENCRYPTED]", vars[tt.args.name])
			}
		})
	}
}
