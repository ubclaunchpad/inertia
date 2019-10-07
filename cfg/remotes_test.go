package cfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var exampleRemote = &Remote{
	Name: "staging",
	IP:   "0.0.0.0",
	SSH: &SSH{
		User: "bob",
	},
	Daemon: &Daemon{
		Port: "4043",
	},
	Profiles: map[string]string{
		"fdsa":  "fdsaf",
		"wqrte": "erterh",
	},
}

func TestNewInertiaConfig(t *testing.T) {
	assert.NotNil(t, NewRemotesConfig())
}

func TestRemotes_GetRemote(t *testing.T) {
	type fields struct {
		Remotes []*Remote
	}
	type args struct {
		name string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *Remote
		wantFound bool
	}{
		{"invalid arg", fields{[]*Remote{exampleRemote}}, args{""}, nil, false},
		{"remote not found", fields{[]*Remote{exampleRemote}}, args{"eh"}, nil, false},
		{"no remotes", fields{[]*Remote{}}, args{"eh"}, nil, false},
		{"find remote",
			fields{[]*Remote{exampleRemote}},
			args{exampleRemote.Name},
			exampleRemote,
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i = &Remotes{
				Remotes: tt.fields.Remotes,
			}
			got, found := i.GetRemote(tt.args.name)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantFound, found)
		})
	}
}

func TestInertia_SetRemote(t *testing.T) {
	type fields struct {
		Remotes []*Remote
	}
	type args struct {
		remote Remote
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Remote
	}{
		{"invalid remote",
			fields{[]*Remote{}},
			args{Remote{}},
			nil},
		{"update remote",
			fields{[]*Remote{exampleRemote}},
			args{Remote{Version: "test", Name: exampleRemote.Name}},
			&Remote{Version: "test", Name: exampleRemote.Name, Daemon: &Daemon{}}},
		{"add new remote",
			fields{[]*Remote{}},
			args{*exampleRemote},
			exampleRemote},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i = &Remotes{
				Remotes: tt.fields.Remotes,
			}
			i.SetRemote(tt.args.remote)
			if tt.want != nil {
				got, found := i.GetRemote(tt.want.Name)
				assert.True(t, found)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestInertia_RemoveRemote(t *testing.T) {
	type fields struct {
		Remotes []*Remote
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
		{"invalid arg", fields{[]*Remote{exampleRemote}}, args{""}, false},
		{"remote not found", fields{[]*Remote{exampleRemote}}, args{"eh"}, false},
		{"no remotes", fields{[]*Remote{}}, args{"eh"}, false},
		{"remove remote",
			fields{[]*Remote{exampleRemote}},
			args{exampleRemote.Name},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i = &Remotes{
				Remotes: tt.fields.Remotes,
			}
			assert.Equal(t, tt.want, i.RemoveRemote(tt.args.name))
			r, found := i.GetRemote(tt.args.name)
			assert.False(t, found)
			assert.Nil(t, r)
		})
	}
}
