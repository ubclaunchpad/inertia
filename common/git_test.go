package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var remoteURLVariations = []string{
	// SSH URL
	"git@github.com:ubclaunchpad/inertia.git",
	"git@gitlab.com:ubclaunchpad/inertia.git",
	"git@bitbucket.org:ubclaunchpad/inertia.git",

	// Github URL Variations
	"https://github.com/ubclaunchpad/inertia.git",
	"git://github.com/ubclaunchpad/inertia.git",

	// Gitlab URL Variations
	"https://gitlab.com/ubclaunchpad/inertia.git",
	"git://gitlab.com/ubclaunchpad/inertia.git",

	// Bitbucket URL Variations
	"https://ubclaunchpad@bitbucket.org/ubclaunchpad/inertia.git",
}

func TestGetSSHRemoteURL(t *testing.T) {
	validSSH := remoteURLVariations[0:3]
	for _, url := range remoteURLVariations {
		assert.Contains(t, validSSH, GetSSHRemoteURL(url))
	}
}

func TestGetBranchFromRef(t *testing.T) {
	type args struct {
		ref string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"single-part branch", args{"refs/heads/master"}, "master"},
		{"two-part branch", args{"refs/heads/web/navbar-hide"}, "web/navbar-hide"},
		{"three-part branch", args{"refs/heads/web/fix/navbar-hide"}, "web/fix/navbar-hide"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBranchFromRef(tt.args.ref); got != tt.want {
				t.Errorf("GetBranchFromRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractRepository(t *testing.T) {
	for _, url := range remoteURLVariations {
		repoName := ExtractRepository(url)
		assert.Equal(t, "ubclaunchpad/inertia", repoName)
	}

	repoNameWithHyphens := ExtractRepository("git@github.com:ubclaunchpad/inertia-deploy-test.git")
	assert.Equal(t, "ubclaunchpad/inertia-deploy-test", repoNameWithHyphens)

	repoNameWithDots := ExtractRepository("git@github.com:ubclaunchpad/inertia.deploy.test.git")
	assert.Equal(t, "ubclaunchpad/inertia.deploy.test", repoNameWithDots)

	repoNameWithMixed := ExtractRepository("git@github.com:ubclaunchpad/inertia-deploy.test.git")
	assert.Equal(t, "ubclaunchpad/inertia-deploy.test", repoNameWithMixed)

	defaultRepoName := ExtractRepository("")
	assert.Equal(t, "${repository}", defaultRepoName)
}
