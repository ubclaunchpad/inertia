package common

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testPrivateKey = []byte("very_sekrit_key")
	testToken      = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.AqFWnFeY9B8jj7-l3z0a9iaZdwIca7xhUF3fuaJjU90"
)

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

func TestFlushOutput(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		reader, writer := io.Pipe()
		go FlushOutput(w, reader)

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

func TestPipeErr(t *testing.T) {
	var b bytes.Buffer
	w := httptest.NewRecorder()
	PipeErr(&b, "Wee!", 200)
	PipeErr(w, "Wee!", 200)
	assert.Equal(t, "[ERROR 200] Wee!", b.String())

	body, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Wee!\n", string(body))
	assert.Equal(t, 200, w.Code)
}

func TestPipeSuccess(t *testing.T) {
	var b bytes.Buffer
	w := httptest.NewRecorder()
	PipeSuccess(&b, "Wee!", 200)
	PipeSuccess(w, "Wee!", 200)
	assert.Equal(t, "Wee!\n", b.String())

	body, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Wee!\n", string(body))
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
}
