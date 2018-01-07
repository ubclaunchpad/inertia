// Copyright © 2017 UBC Launch Pad team@ubclaunchpad.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	malformedAuthStringErrorMsg = "Malformed authentication string"
	noAuthTokenErrorMsg         = "Must provide auth token"
	tokenInvalidErrorMsg        = "Token invalid"
)

// checkForGit returns an error if we're not in a git repository.
func checkForGit(cwd string) error {
	// Quick failure if no .git folder.
	gitFolder := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitFolder); os.IsNotExist(err) {
		return errors.New("this does not appear to be a git repository")
	}

	repo, err := git.PlainOpen(cwd)
	if err != nil {
		return err
	}

	remotes, err := repo.Remotes()

	// Also fail if no remotes detected.
	if len(remotes) == 0 {
		return errors.New("there are no remotes associated with this repository")
	}

	return nil
}

// getLocalRepo gets the repo from disk.
func getLocalRepo() (*git.Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return git.PlainOpen(cwd)
}

// getSSHRemoteURL gets the URL of the given remote in the form
// "git@github.com:[USER]/[REPOSITORY].git"
func getSSHRemoteURL(url string) string {
	newURL := strings.Replace(url, "https://github.com/", "git@github.com:", -1)
	if url != newURL {
		return newURL + ".git"
	}
	return url
}

// removeContents removes all files within given directory, returns nil if successful
func removeContents(directory string) error {
	d, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(directory, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// authorized is a function decorator for authorizing RESTful
// daemon requests. It wraps handler functions and ensures the
// request is authorized. Returns a function
func authorized(handler http.HandlerFunc, keyLookup func(*jwt.Token) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Collect the token from the header.
		bearerString := r.Header.Get("Authorization")

		// Split out the actual token from the header.
		splitToken := strings.Split(bearerString, "Bearer ")
		if len(splitToken) < 2 {
			http.Error(w, malformedAuthStringErrorMsg, http.StatusForbidden)
			return
		}
		tokenString := splitToken[1]

		// Parse takes the token string and a function for looking up the key.
		token, err := jwt.Parse(tokenString, keyLookup)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// Verify the claims (none for now) and token.
		if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
			http.Error(w, tokenInvalidErrorMsg, http.StatusForbidden)
			return
		}

		// We're authorized, run the handler.
		handler(w, r)
	}
}

// generateToken creates a JSON Web Token (JWT) for a client to use when
// sending HTTP requests to the daemon server.
func generateToken(key []byte) (string, error) {
	// No claims for now.
	return jwt.New(jwt.SigningMethodHS256).SignedString(key)
}

// getAPIPrivateKey returns the private RSA key to authenticate HTTP
// requests sent to the daemon. For now, we simply use the GitHub
// deploy key.
func getAPIPrivateKey(*jwt.Token) (interface{}, error) {
	pemFile, err := os.Open(daemonGithubKeyLocation)
	if err != nil {
		return nil, err
	}
	key, err := getGithubKey(pemFile)
	if err != nil {
		return nil, err
	}
	return []byte(key.String()), nil
}

// getAPIPrivateKeyFromDaemon returns the private RSA key to authenticate HTTP
// requests sent to the daemon from local config. For now, we simply use the GitHub
// deploy key.
func getAPIPrivateKeyFromConfig() (string, error) {
	cfg, err := getProjectConfigFromDisk()
	if err != nil {
		return "", err
	}
	return cfg.DaemonAPIToken, nil
}

// getGithubKey returns an ssh.AuthMethod from the given io.Reader
// for use with the go-git library
func getGithubKey(pemFile io.Reader) (ssh.AuthMethod, error) {
	bytes, err := ioutil.ReadAll(pemFile)
	if err != nil {
		return nil, err
	}
	return ssh.NewPublicKeys("git", bytes, "")
}

// clone wraps git.PlainClone() and returns a more helpful error message
// if the given error is an authentication-related error.
func clone(directory string, remoteURL string, auth ssh.AuthMethod) (*git.Repository, error) {
	repo, err := git.PlainClone(projectDirectory, false, &git.CloneOptions{
		URL:  remoteURL,
		Auth: auth,
	})
	if err == transport.ErrInvalidAuthMethod || err == transport.ErrAuthenticationRequired || err == transport.ErrAuthorizationFailed {
		return nil, errors.New("Access to project repository rejected; did you forget to add\nInertia's deploy key to your repository settings?\n" + auth.String())
	} else if err != nil {
		return nil, err
	}
	return repo, nil
}

// forcePull deletes the project directory and makes a fresh clone of given repo
// git.Worktree.Pull() only supports merges that can be resolved as a fast-forward
func forcePull(repo *git.Repository, auth ssh.AuthMethod) (*git.Repository, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}
	remoteURL := getSSHRemoteURL(remotes[0].Config().URLs[0])
	err = removeContents(projectDirectory)
	if err != nil {
		return nil, err
	}
	repo, err = clone(projectDirectory, remoteURL, auth)
	if err != nil {
		e := removeContents(projectDirectory)
		if e != nil {
			log.WithError(e)
		}
		return nil, err
	}
	return repo, nil
}

// compareRemotes checks if the given remote matches the remote of the given repository
func compareRemotes(localRepo *git.Repository, remoteURL string) error {
	remotes, err := localRepo.Remotes()
	if err != nil {
		return err
	}
	localRemoteURL := getSSHRemoteURL(remotes[0].Config().URLs[0])
	if localRemoteURL != remoteURL {
		return errors.New("The given remote URL does not match that of the repository in\nyour remote - try 'inertia deploy [REMOTE] reset'")
	}
	return nil
}
