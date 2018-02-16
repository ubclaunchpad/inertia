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

package common

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// CheckForGit returns an error if we're not in a git repository.
func CheckForGit(cwd string) error {
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
	if err != nil {
		return err
	}

	// Also fail if no remotes detected.
	if len(remotes) == 0 {
		return errors.New("there are no remotes associated with this repository")
	}

	return nil
}

// CheckForDockerCompose returns error if current directory is a
// not a docker-compose project
func CheckForDockerCompose(cwd string) error {
	dockerComposeYML := filepath.Join(cwd, "docker-compose.yml")
	dockerComposeYAML := filepath.Join(cwd, "docker-compose.yaml")
	_, err := os.Stat(dockerComposeYML)
	YMLpresent := os.IsNotExist(err)
	_, err = os.Stat(dockerComposeYAML)
	YAMLpresent := os.IsNotExist(err)
	if YMLpresent && YAMLpresent {
		return errors.New("this does not appear to be a docker-compose project - currently,\n" +
			"Inertia only supports docker-compose projects.")
	}
	return nil
}

// GetLocalRepo gets the repo from disk.
func GetLocalRepo() (*git.Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return git.PlainOpen(cwd)
}

// GetSSHRemoteURL gets the URL of the given remote in the form
// "git@github.com:[USER]/[REPOSITORY].git"
func GetSSHRemoteURL(url string) string {
	return strings.Replace(url, "https://github.com/", "git@github.com:", -1)
}

// RemoveContents removes all files within given directory, returns nil if successful
func RemoveContents(directory string) error {
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

// GenerateToken creates a JSON Web Token (JWT) for a client to use when
// sending HTTP requests to the daemon server.
func GenerateToken(key []byte) (string, error) {
	// No claims for now.
	return jwt.New(jwt.SigningMethodHS256).SignedString(key)
}

// GetGithubKey returns an ssh.AuthMethod from the given io.Reader
// for use with the go-git library
func GetGithubKey(pemFile io.Reader) (ssh.AuthMethod, error) {
	bytes, err := ioutil.ReadAll(pemFile)
	if err != nil {
		return nil, err
	}
	return ssh.NewPublicKeys("git", bytes, "")
}

// Clone wraps git.PlainClone() and returns a more helpful error message
// if the given error is an authentication-related error.
func Clone(directory string, remoteURL string, auth ssh.AuthMethod, out io.Writer) (*git.Repository, error) {
	repo, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:      remoteURL,
		Auth:     auth,
		Depth:    2,
		Progress: out,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, err
	}

	// Use this to confirm if pull has completed.
	_, err = repo.Head()
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// ForcePull deletes the project directory and makes a fresh clone of given repo
// git.Worktree.Pull() only supports merges that can be resolved as a fast-forward
func ForcePull(directory string, repo *git.Repository, auth ssh.AuthMethod, out io.Writer) (*git.Repository, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}
	remoteURL := GetSSHRemoteURL(remotes[0].Config().URLs[0])
	err = RemoveContents(directory)
	if err != nil {
		return nil, err
	}
	repo, err = Clone(directory, remoteURL, auth, out)
	if err != nil {
		e := RemoveContents(directory)
		if e != nil {
			log.WithError(e)
		}
		return nil, err
	}
	return repo, nil
}

// CompareRemotes checks if the given remote matches the remote of the given repository
func CompareRemotes(localRepo *git.Repository, remoteURL string) error {
	remotes, err := localRepo.Remotes()
	if err != nil {
		return err
	}
	localRemoteURL := GetSSHRemoteURL(remotes[0].Config().URLs[0])
	if localRemoteURL != remoteURL {
		return errors.New("The given remote URL does not match that of the repository in\nyour remote - try 'inertia [REMOTE] reset'")
	}
	return nil
}

// FlushOutput continuously writes everything in given PipeReader
// to a ResponseWriter. Use this as a goroutine.
func FlushOutput(w io.Writer, pipeReader io.ReadCloser) {
	buffer := make([]byte, 100)
	for {
		// Read from pipe then write to ResponseWriter and flush it,
		// sending the copied content to the client.
		n, err := pipeReader.Read(buffer)
		if err != nil {
			pipeReader.Close()
			break
		}
		data := buffer[0:n]
		w.Write(data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Clear the buffer.
		for i := 0; i < n; i++ {
			buffer[i] = 0
		}
	}
}

// PipeErr directs message and status to http.Error when appropriate,
// otherwise pipes to the given io.Writer using fmt.Print
func PipeErr(w io.Writer, msg string, status int) {
	responseWriter, ok := w.(http.ResponseWriter)
	if ok {
		http.Error(responseWriter, msg, status)
	} else {
		fmt.Fprintf(w, "[ERROR %s] %s", strconv.Itoa(status), msg)
	}
}

// PipeSuccess directs status to Header and sets content type when appropriate
// and writes message to given io.Writer
func PipeSuccess(w io.Writer, msg string, status int) {
	responseWriter, ok := w.(http.ResponseWriter)
	if ok {
		responseWriter.Header().Set("Content-Type", "text/html")
		responseWriter.WriteHeader(status)
	}
	fmt.Fprintln(w, msg)
}
