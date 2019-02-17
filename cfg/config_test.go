package cfg

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	type args struct {
		path    string
		data    interface{}
		writers []io.Writer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"nothing to write to", args{"", nil, nil}, true},
		{"ok: write to path", args{"./test-config.toml", &Inertia{
			Remotes: make(map[string]Remote),
		}, nil}, false},
		{"ok: write to writers", args{"", &Inertia{
			Remotes: make(map[string]Remote),
		}, []io.Writer{os.Stdout}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.path != "" {
				defer os.RemoveAll(tt.args.path)
			}
			var err = Write(tt.args.path, tt.args.data, tt.args.writers...)
			assert.Equalf(t, (err != nil), tt.wantErr, "got '%v'", err)
		})
	}
}
