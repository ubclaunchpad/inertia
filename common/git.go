package common

import (
	"regexp"
	"strings"
)

// GetSSHRemoteURL gets the URL of the given remote in the form
// "git@github.com:[USER]/[REPOSITORY].git"
func GetSSHRemoteURL(url string) string {
	re, _ := regexp.Compile("(https|git)://")

	sshURL := re.ReplaceAllString(url, "git@")
	if sshURL != url {
		sshURL = strings.Replace(sshURL, "/", ":", 1)
	}

	// special bitbucket https case?
	lastIndex := strings.LastIndex(sshURL, "@")
	if lastIndex > 3 {
		sshURL = "git@" + sshURL[lastIndex+1:]
	}

	if !strings.HasSuffix(sshURL, ".git") {
		sshURL = sshURL + ".git"
	}
	return sshURL
}

// GetBranchFromRef gets the branch name from a git ref of form refs/...
func GetBranchFromRef(ref string) string {
	parts := strings.Split(ref, "/")
	return strings.Join(parts[2:], "/")
}

// ExtractRepository gets the project name from its URL in the form [username]/[project]
func ExtractRepository(URL string) string {
	re, err := regexp.Compile(":|/")
	if err != nil {
		return "$YOUR_REPOSITORY"
	}
	var parts = re.Split(strings.TrimSuffix(URL, ".git"), -1)
	if len(parts) < 2 {
		return "${repository}"
	}
	return strings.Join(parts[len(parts)-2:], "/")
}
