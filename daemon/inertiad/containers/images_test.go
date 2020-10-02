package containers

import (
	"context"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
)

var testVersion = semver.MustParse("0.6.1")

func semverMustParsePtr(v string) *semver.Version {
	sv := semver.MustParse(v)
	return &sv
}

func Test_imageTagsResult_getLatest(t *testing.T) {
	type args struct {
		min *semver.Version
	}
	tests := []struct {
		name    string
		fields  dockerHubImageTagsResult
		args    args
		want    *semver.Version
		wantErr bool
	}{
		{"should get latest if no min is provided",
			dockerHubImageTagsResult{Results: []dockerHubImageTagDescription{
				{Name: "v0.7.0"}, {Name: "v0.6.1"}, {Name: "v0.6.0-rc1"}, {Name: "v0.6.0-preview1"},
			}},
			args{},
			semverMustParsePtr("0.7.0"),
			false},
		{"should get latest if min is provided",
			dockerHubImageTagsResult{Results: []dockerHubImageTagDescription{
				{Name: "v0.7.0"}, {Name: "v0.6.1"}, {Name: "v0.6.0-rc1"}, {Name: "v0.6.0-preview1"},
			}},
			args{&testVersion},
			semverMustParsePtr("0.7.0"),
			false},
		{"should return same version if nothing newer is available",
			dockerHubImageTagsResult{Results: []dockerHubImageTagDescription{
				{Name: "v0.6.0-rc1"}, {Name: "v0.6.0-preview1"},
			}},
			args{&testVersion},
			&testVersion,
			false},
		{"error if no new is available",
			dockerHubImageTagsResult{Results: []dockerHubImageTagDescription{}},
			args{},
			nil,
			true},
		{"should NOT return release candidates",
			dockerHubImageTagsResult{Results: []dockerHubImageTagDescription{
				{Name: "v0.6.0-rc1"}, {Name: "v0.5.0"},
			}},
			args{},
			semverMustParsePtr("0.5.0"),
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.getVersions().getLatest(tt.args.min)
			if (err != nil) != tt.wantErr {
				t.Errorf("%+v", got)
				t.Errorf("imageTagsResult.getLatest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equals(*tt.want) {
				t.Errorf("imageTagsResult.getLatest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLatestImageTag(t *testing.T) {
	v := semver.MustParse("0.5.0")
	latest, err := GetLatestImageTag(context.Background(), Image{
		Registry:   "ghcr.io",
		Repository: "ubclaunchpad/inertia",
	}, &v)
	assert.NoError(t, err)
	assert.NotNil(t, latest)
	assert.True(t, latest.GT(v))

	latest, err = GetLatestImageTag(context.Background(), Image{
		Repository: "docker/compose",
	}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, latest)
}
