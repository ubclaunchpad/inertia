// Proof-of-concept Github Webhook receiver for receiving POST requests from
// Github and determining the event, branch, and repository from the request
// payload.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/go-github/github"
)

var secret = ""

// Arguments: [port] [secret]
// Secret is the secret key used to secure the HTTP POST payload. You enter this
// secret when you're setting up your Github Webhook.
func main() {
	// Process CLI args
	if len(os.Args) != 3 {
		log.Println("Usage: webhooks [port] [secret]")
		return
	}

	port, err := strconv.ParseInt(os.Args[1], 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Invalid port: %s", err))
		return
	}
	secret = os.Args[2]

	log.Println(fmt.Sprintf("Starting POC Webhook Receiver on port %v", port))

	// Route all requests to the requestHandler.
	http.HandleFunc("/", requestHandler)

	// Listen on port 8081. Pass nil as handler so TCP connections are handled
	// by the DefaultServeMux.
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

// requestHandler writes a response to a request into the given ResponseWriter.
func requestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "I'm a little Webhook, short and stout!")

	payload, err := github.ValidatePayload(r, []byte(secret))
	if err != nil {
		log.Println(err)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Println(err)
		return
	}

	switch event := event.(type) {
	case *github.PushEvent:
		processPushEvent(event)
	case *github.PullRequestEvent:
		processPullRequestEvent(event)
	default:
		log.Println("Unrecognized event type")
	}
}

// processPushEvent prints information about the given PushEvent.
func processPushEvent(event *github.PushEvent) {
	repo := event.GetRepo()
	log.Println("Received PushEvent")
	log.Println(fmt.Sprintf("Repository Name: %s", *repo.Name))
	log.Println(fmt.Sprintf("Repository Git URL: %s", *repo.GitURL))
	log.Println(fmt.Sprintf("Ref: %s", event.GetRef()))
}

// processPullREquestEvent prints information about the given PullRequestEvent.
// Handling PRs is unnecessary because merging one will trigger a PushEvent.
func processPullRequestEvent(event *github.PullRequestEvent) {
	repo := event.GetRepo()
	pr := event.GetPullRequest()
	merged := "false"
	if *pr.Merged {
		merged = "true"
	}
	log.Println("Received PullRequestEvent")
	log.Println(fmt.Sprintf("Repository Name: %s", *repo.Name))
	log.Println(fmt.Sprintf("Repository Git URL: %s", *repo.GitURL))
	log.Println(fmt.Sprintf("Ref: %s", pr.GetBase().GetRef()))
	log.Println(fmt.Sprintf("Merge status: %v", merged))
}
