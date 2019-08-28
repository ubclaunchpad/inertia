package cfg

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetProperty(t *testing.T) {
	type args struct {
		property string
		value    string
		obj      interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		verifyFunc func(interface{}) error
	}{
		{"ok: set string",
			args{"version", "v0.5.2", &Remote{}},
			false,
			func(d interface{}) error {
				var remote = d.(*Remote)
				if remote.Version != "v0.5.2" {
					return fmt.Errorf("value not set (found '%s')", remote.Version)
				}
				return nil
			}},
		{"ok: set nested string",
			args{"daemon.port", "8000", &Remote{
				Daemon: &Daemon{Port: "8080"},
			}},
			false,
			func(d interface{}) error {
				var remote = d.(*Remote)
				if remote.Daemon.Port != "8000" {
					return fmt.Errorf("value not set (found '%s')", remote.Daemon.Port)
				}
				return nil
			}},
		{"ok: set nested typed string",
			args{"build.type", "dockerfile", &Profile{
				Build: &Build{Type: DockerCompose},
			}},
			false,
			func(d interface{}) error {
				var profile = d.(*Profile)
				if profile.Build.Type != Dockerfile {
					return errors.New("value not set")
				}
				return nil
			}},
		{"ok: set boolean",
			args{"daemon.verify-ssl", "true", &Remote{
				Daemon: &Daemon{VerifySSL: false},
			}},
			false,
			func(d interface{}) error {
				var remote = d.(*Remote)
				if !remote.Daemon.VerifySSL {
					return fmt.Errorf("value not set (found '%t')", remote.Daemon.VerifySSL)
				}
				return nil
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err = SetProperty(tt.args.property, tt.args.value, tt.args.obj)
			assert.Equalf(t, (err != nil), tt.wantErr, "got '%+v'", err)
			assert.NoError(t, tt.verifyFunc(tt.args.obj))
		})
	}
}
