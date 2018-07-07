package webhook

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ubclaunchpad/inertia/common"
)

func parse(r *http.Request) (string, error) {
	fmt.Println("Parsing webhook...")

	if r.Header.Get("content-type") != "application/json" {
		return "", errors.New("Content-Type must be JSON")
	}

	// Try Github
	githubEventHeader := r.Header.Get("x-github-event")
	if len(githubEventHeader) > 0 {
		return "Github webhook received", nil
	}

	// Try Gitlab
	gitlabEventHeader := r.Header.Get("x-gitlab-event")
	if len(gitlabEventHeader) > 0 {
		return "Gitlab webhook received", nil
	}

	// Try Bitbucket
	userAgent := r.Header.Get("user-agent")
	if strings.Contains(userAgent, "BitBucket") {
		return "Bitbucket webhook received", nil
	}

	return "", errors.New("Unsupported webhook received")
}

// Handler receives a webhook and parses it into one of the supported types
func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, common.MsgDaemonOK)
	payload, err := parse(r)
	if err != nil {
		println(err)
		return
	}

	println(payload)
}
