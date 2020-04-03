package cfg

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var exampleProfile = &Profile{
	Name:   "test",
	Branch: "master",
	Build: &Build{
		Type:          DockerCompose,
		BuildFilePath: "docker-compose.yml",
	},
}

func TestNewProject(t *testing.T) {
	assert.NotNil(t, NewProject("test", "blah.git"))
	assert.NotNil(t, NewProject("", "blah.git"))
}

func TestProject_GetProfile(t *testing.T) {
	type fields struct {
		Profiles []*Profile
	}
	type args struct {
		name string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *Profile
		wantFound bool
	}{
		{"invalid arg", fields{[]*Profile{exampleProfile}}, args{""}, nil, false},
		{"profile not found", fields{[]*Profile{exampleProfile}}, args{"eh"}, nil, false},
		{"no profiles", fields{[]*Profile{}}, args{"eh"}, nil, false},
		{"find profile",
			fields{[]*Profile{exampleProfile}},
			args{exampleProfile.Name},
			exampleProfile,
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p = &Project{
				Profiles: tt.fields.Profiles,
			}
			got, found := p.GetProfile(tt.args.name)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantFound, found)
		})
	}
}

func TestProject_SetProfile(t *testing.T) {
	type fields struct {
		Profiles []*Profile
	}
	type args struct {
		profile Profile
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Profile
	}{
		{"invalid profile",
			fields{[]*Profile{}},
			args{Profile{}},
			nil},
		{"update profile",
			fields{[]*Profile{exampleProfile}},
			args{Profile{Name: exampleProfile.Name, Branch: "wow"}},
			&Profile{Name: exampleProfile.Name, Branch: "wow", Build: &Build{}}},
		{"add new profile",
			fields{[]*Profile{}},
			args{*exampleProfile},
			exampleProfile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p = &Project{
				Profiles: tt.fields.Profiles,
			}
			p.SetProfile(tt.args.profile)
			if tt.want != nil {
				got, found := p.GetProfile(tt.want.Name)
				assert.True(t, found)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestProject_RemoveProfile(t *testing.T) {
	type fields struct {
		Profiles []*Profile
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"invalid arg", fields{[]*Profile{exampleProfile}}, args{""}, false},
		{"profile not found", fields{[]*Profile{exampleProfile}}, args{"eh"}, false},
		{"no profile", fields{[]*Profile{}}, args{"eh"}, false},
		{"remove profile",
			fields{[]*Profile{exampleProfile}},
			args{exampleProfile.Name},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p = &Project{
				Profiles: tt.fields.Profiles,
			}
			assert.Equal(t, tt.want, p.RemoveProfile(tt.args.name))
			profile, found := p.GetProfile(tt.args.name)
			assert.False(t, found)
			assert.Nil(t, profile)
		})
	}
}

func TestProject_ValidateVersion(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name    string
		fields  *Project
		args    args
		want    string
		wantErr bool
	}{
		{"no version provided",
			&Project{InertiaMinVersion: "v0.5.0"},
			args{""},
			"",
			true},
		{"no version configured",
			&Project{InertiaMinVersion: ""},
			args{"v0.5.0"},
			"no inertia version",
			false},
		{"not in range",
			&Project{InertiaMinVersion: "v0.5.3"},
			args{"v0.6.0"},
			"",
			true},
		{"ok - same version",
			&Project{InertiaMinVersion: "v0.5.3"},
			args{"v0.5.3"},
			"",
			false},
		{"ok - higher version",
			&Project{InertiaMinVersion: "v0.5.3"},
			args{"v0.5.8"},
			"",
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.ValidateVersion(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Project.ValidateVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("Project.ValidateVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
