package cfg

import (
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
			if got := p.RemoveProfile(tt.args.name); got != tt.want {
				t.Errorf("Inertia.RemoveRemote() = %v, want %v", got, tt.want)
			}
			profile, found := p.GetProfile(tt.args.name)
			assert.False(t, found)
			assert.Nil(t, profile)
		})
	}
}
