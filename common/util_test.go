package common

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/src-d/go-git.v4/config"
)

var (
	testPrivateKey = []byte("very_sekrit_key")
	testToken      = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.AqFWnFeY9B8jj7-l3z0a9iaZdwIca7xhUF3fuaJjU90"
)

func getMockRepo(url string) (*git.Repository, error){
	memory := memory.NewStorage()

	if url[len(url) - len(".git"):] != ".git" {
		return nil, fmt.Errorf("the given URL '%s' must end in '.git'", url)
	}

	mockRepo, err := git.Init(memory, nil)
	if err != nil {
		return nil, err
	}
	_, err = mockRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	if err != nil {
		return nil, err
	}
	return mockRepo, nil
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(testPrivateKey)
	assert.Nil(t, err, "generateToken must not fail")
	assert.Equal(t, token, testToken)
}

func TestGetGithubKey(t *testing.T) {
	pemFile, err := os.Open(os.Getenv("GOPATH") + "/src/github.com/ubclaunchpad/inertia/test_env/test_key")
	assert.Nil(t, err)
	_, err = GetGithubKey(pemFile)
	assert.Nil(t, err)
}

func TestGetSSHRemoteURL(t *testing.T) {
	httpsURL := "https://github.com/ubclaunchpad/inertia.git"
	sshURL := "git@github.com:ubclaunchpad/inertia.git"

	assert.Equal(t, sshURL, GetSSHRemoteURL(httpsURL))
	assert.Equal(t, sshURL, GetSSHRemoteURL(sshURL))
}

func TestCheckForGit(t *testing.T) {
	cwd, _ := os.Getwd()
	assert.NotEqual(t, nil, CheckForGit(cwd))
	inertia := strings.TrimSuffix(cwd, "/common")
	assert.Equal(t, nil, CheckForGit(inertia))
}

func TestCheckForDockerCompose(t *testing.T) {
	cwd, _ := os.Getwd()
	assert.NotEqual(t, nil, CheckForDockerCompose(cwd))
	file, _ := os.Create(cwd + "/docker-compose.yml")
	file.Close()
	assert.Equal(t, nil, CheckForDockerCompose(cwd))
	os.Remove(cwd + "/docker-compose.yml")
	file, _ = os.Create(cwd + "/docker-compose.yaml")
	file.Close()
	assert.Equal(t, nil, CheckForDockerCompose(cwd))
	os.Remove(cwd + "/docker-compose.yaml")
}

func TestFlushRoutine(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		reader, writer := io.Pipe()
		go FlushRoutine(w, reader)

		fmt.Println(writer, "Hello!")
		time.Sleep(time.Millisecond)

		fmt.Println(writer, "Lunch?")
		time.Sleep(time.Millisecond)

		fmt.Println(writer, "Bye!")
		time.Sleep(time.Millisecond)
	}))
	defer testServer.Close()

	resp, err := http.DefaultClient.Get(testServer.URL)
	assert.Nil(t, err)

	reader := bufio.NewReader(resp.Body)
	i := 0
	for i < 3 {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		switch i {
		case 0:
			assert.Equal(t, "Hello!", string(line))
		case 1:
			assert.Equal(t, "Lunch?", string(line))
		case 2:
			assert.Equal(t, "Bye!", string(line))
		}

		i++
	}
}

func TestGetProjectName(t *testing.T) {
	testRemote1, err := getMockRepo("https://github.com/ubclaunchpad/inertia.git")
	assert.Nil(t, err)
	repoName, err := GetProjectName(testRemote1)
	assert.Equal(t, "inertia", repoName)

	a, err := getMockRepo("https://github.com/ubclaunchpad/inertia")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(a)
	}
}
