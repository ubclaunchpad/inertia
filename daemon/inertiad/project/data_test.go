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
		{"invalid env", args{"", "", true}, true, true},
		{"no encrypt", args{"myvar1", "mysekret", false}, true, false},
		{"encrypt", args{"myvar2", "myothersekret", true}, true, false},
		{"no decrypt", args{"myvar", "asdfasdf", true}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := "./test_config"
			err := os.Mkdir(dir, os.ModePerm)
			assert.Nil(t, err)
			defer os.RemoveAll(dir)

			// Instantiate
			c, err := newDataManager(path.Join(dir, "deployment.db"))
			assert.Nil(t, err)

			// Add
			err = c.AddEnvVariable(tt.args.name, tt.args.value, tt.args.encrypt)
			assert.Equal(t, tt.wantErr, (err != nil))

			// Retrieve
			vars, err := c.GetEnvVariables(tt.decrypt)
			assert.Nil(t, err)
			if tt.wantErr {
				assert.Zero(t, len(vars))
			} else {
				if tt.decrypt {
					assert.Equal(t, tt.args.name+"="+tt.args.value, vars[0])
				} else {
					assert.Equal(t, tt.args.name+"=[ENCRYPTED]", vars[0])
				}
			}

			// Remove
			err = c.RemoveEnvVariables(tt.args.name)
			assert.Nil(t, err)
			vars, err = c.GetEnvVariables(false)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(vars))
		})
	}
}

func TestDataManager_destroy(t *testing.T) {
	dir := "./test_config"
	err := os.Mkdir(dir, os.ModePerm)
	assert.Nil(t, err)
	defer os.RemoveAll(dir)

	// Instantiate
	c, err := newDataManager(path.Join(dir, "deployment.db"))
	assert.Nil(t, err)

	// Reset
	err = c.destroy()
	assert.Nil(t, err)

	// Check if bucket is still usable
	_, err = c.GetEnvVariables(false)
	assert.Nil(t, err)
}
