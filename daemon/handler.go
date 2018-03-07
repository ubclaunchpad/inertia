package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/google/go-github/github"
	git "gopkg.in/src-d/go-git.v4"

	"github.com/ubclaunchpad/inertia/common"
)

const (
	// specify location of deployed project
	projectDirectory = "/app/host/project"

	// specify docker-compose version
	dockerCompose = "docker/compose:1.18.0"

	// specify common responses here
	noContainersResp            = "There are currently no active containers."
	malformedAuthStringErrorMsg = "Malformed authentication string"
	tokenInvalidErrorMsg        = "Token invalid"

	defaultSecret = "inertia"
)

// daemonVersion indicates the daemon's corresponding Inertia daemonVersion
var daemonVersion string

// Run starts the daemon
func Run(port, version string) {
	daemonVersion = version

	// Download docker-compose image
	println("Downloading docker-compose...")
	cli, err := docker.NewEnvClient()
	if err != nil {
		println(err)
		println("Failed to start Docker client - shutting down daemon.")
		return
	}
	_, err = cli.ImagePull(context.Background(), dockerCompose, types.ImagePullOptions{})
	if err != nil {
		println(err)
		println("Failed to pull docker-compose image - shutting down daemon.")
		cli.Close()
		return
	}
	cli.Close()

	// Run daemon on port
	println("Serving daemon on port " + port)
	mux := http.NewServeMux()
	mux.HandleFunc("/health-check", authorized(healthCheckHandler, GetAPIPrivateKey))
	mux.HandleFunc("/", gitHubWebHookHandler)
	mux.HandleFunc("/up", authorized(upHandler, GetAPIPrivateKey))
	mux.HandleFunc("/down", authorized(downHandler, GetAPIPrivateKey))
	mux.HandleFunc("/status", authorized(statusHandler, GetAPIPrivateKey))
	mux.HandleFunc("/reset", authorized(resetHandler, GetAPIPrivateKey))
	mux.HandleFunc("/logs", authorized(logHandler, GetAPIPrivateKey))
	print(http.ListenAndServe(":"+port, mux))
}

// healthCheckHandler returns a 200 if the daemon is happy.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, common.DaemonOkResp)
}

// gitHubWebHookHandler writes a response to a request into the given ResponseWriter.
func gitHubWebHookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, common.DaemonOkResp)

	payload, err := github.ValidatePayload(r, []byte(defaultSecret))
	if err != nil {
		println(err)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		println(err)
		return
	}

	switch event := event.(type) {
	case *github.PushEvent:
		processPushEvent(event)
	case *github.PullRequestEvent:
		processPullRequestEvent(event)
	default:
		println("Unrecognized event type")
	}
}

// upHandler tries to bring the deployment online
func upHandler(w http.ResponseWriter, r *http.Request) {
	println("UP request received")

	// Get github URL from up request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var upReq common.DaemonRequest
	err = json.Unmarshal(body, &upReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger := newLogger(upReq.Stream, w)
	gitOpts := upReq.GitOptions
	defer logger.Close()

	// Check for existing git repository, clone if no git repository exists.
	err = common.CheckForGit(projectDirectory)
	if err != nil {
		logger.Println("No git repository present.")
		err = setUpProject(gitOpts.RemoteURL, gitOpts.Branch, logger.GetWriter())
		if err != nil {
			logger.Err(err.Error(), http.StatusPreconditionFailed)
			return
		}
	}

	repo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		logger.Err(err.Error(), http.StatusPreconditionFailed)
		return
	}

	// Check for matching remotes
	err = common.CompareRemotes(repo, gitOpts.RemoteURL)
	if err != nil {
		logger.Err(err.Error(), http.StatusPreconditionFailed)
		return
	}

	// Update and deploy project
	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	err = deploy(repo, gitOpts.Branch, cli, logger.GetWriter())
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Success("Project startup initiated!", http.StatusCreated)
}

// downHandler tries to take the deployment offline
func downHandler(w http.ResponseWriter, r *http.Request) {
	println("DOWN request received")

	logger := newLogger(false, w)
	defer logger.Close()

	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	// Error if no project containers are active, but try to kill
	// everything anyway in case the docker-compose image is still
	// active
	_, err = getActiveContainers(cli)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		err = killActiveContainers(cli, logger.GetWriter())
		if err != nil {
			println(err)
		}
		return
	}

	err = killActiveContainers(cli, logger.GetWriter())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Project shut down.")
}

// statusHandler lists currently active project containers
func statusHandler(w http.ResponseWriter, r *http.Request) {
	println("STATUS request received")

	inertiaStatus := "inertia daemon " + daemonVersion + "\n"

	// Get status of repository
	repo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		return
	}
	head, err := repo.Head()
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		return
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return
	}
	branchStatus := " - Branch:  " + head.Name().Short() + "\n"
	commitStatus := " - Commit:  " + head.Hash().String() + "\n"
	commitMessage := " - Message: " + commit.Message + "\n"
	status := inertiaStatus + branchStatus + commitStatus + commitMessage

	// Get containers
	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	containers, err := getActiveContainers(cli)
	if err != nil {
		if err.Error() == noContainersResp {
			// This is different from having 2 containers active -
			// noContainersResp means that no attempt to build the project
			// was made or the project was cleanly shut down.
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, status+noContainersResp)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If there are only 2 containers active, that means that a build
	// attempt was made but only the daemon and the docker-compose containers
	// are active, indicating a build failure.
	if len(containers) == 2 {
		errorString := status + "It appears that an attempt to start your project was made but the build failed."
		http.Error(w, errorString, http.StatusNotFound)
		return
	}

	ignore := map[string]bool{
		"/inertia-daemon": true,
		"/docker-compose": true,
	}
	// Only list project containers
	activeContainers := "Active containers:"
	for _, container := range containers {
		if !ignore[container.Names[0]] {
			activeContainers += "\n" + container.Image + " (" + container.Names[0] + ")"
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, status+activeContainers)
}

// resetHandler shuts down and wipes the project directory
func resetHandler(w http.ResponseWriter, r *http.Request) {
	println("RESET request received")

	logger := newLogger(false, w)
	defer logger.Close()

	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	err = killActiveContainers(cli, logger.GetWriter())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = common.RemoveContents(projectDirectory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Project removed from remote.")
}

// logHandler handles requests for container logs
func logHandler(w http.ResponseWriter, r *http.Request) {
	println("LOG request received")

	// Get container name from request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var upReq common.DaemonRequest
	err = json.Unmarshal(body, &upReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	container := upReq.Container
	logger := newLogger(upReq.Stream, w)
	defer logger.Close()

	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	ctx := context.Background()
	logs, err := cli.ContainerLogs(ctx, container, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     upReq.Stream,
		Timestamps: true,
	})
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	defer logs.Close()

	if upReq.Stream {
		common.FlushRoutine(w, logs)
	} else {
		buf := new(bytes.Buffer)
		buf.ReadFrom(logs)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, buf.String())
	}
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
