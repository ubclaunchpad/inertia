package daemon

import (
	"bytes"
	"net/http"
	"os"
	"strconv"
	"strings"

	docker "github.com/docker/docker/client"
	"github.com/go-chi/render"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

// logHandler handles requests for container logs
func (s *Server) logHandler(w http.ResponseWriter, r *http.Request) {
	var (
		shouldStream bool
		err          error
	)

	// Get container name and stream from request query params
	params := r.URL.Query()
	container := params.Get(api.Container)
	streamParam := params.Get(api.Stream)
	if streamParam != "" {
		s, err := strconv.ParseBool(streamParam)
		if err != nil {
			println(err.Error())
			render.Render(w, r, res.ErrBadRequest(err.Error()))
			return
		}
		shouldStream = s
	} else {
		shouldStream = false
	}

	// Determine number of entries to fetch
	entriesParam := params.Get(api.Entries)
	var entries int
	if entriesParam != "" {
		if entries, err = strconv.Atoi(entriesParam); err != nil {
			render.Render(w, r, res.ErrBadRequest("invalid number of entries",
				"error", err))
			return
		}
	}
	if entries == 0 {
		entries = 500
	}

	// Upgrade to websocket connection if required, otherwise just set up a
	// standard streamer
	var stream *log.Streamer
	if shouldStream {
		socket, err := s.websocket.Upgrade(w, r, nil)
		if err != nil {
			render.Render(w, r,
				res.ErrInternalServer("failed to esablish websocket connection", err))
			return
		}
		stream = log.NewStreamer(log.StreamerOptions{
			Request:    r,
			Stdout:     os.Stdout,
			Socket:     socket,
			HTTPWriter: w,
		})
	} else {
		stream = log.NewStreamer(log.StreamerOptions{
			Request:    r,
			Stdout:     os.Stdout,
			HTTPWriter: w,
		})
	}

	logs, err := containers.ContainerLogs(s.docker, containers.LogOptions{
		Container: container,
		Stream:    shouldStream,
		Entries:   entries,
	})
	if err != nil {
		if docker.IsErrNotFound(err) {
			stream.Error(res.ErrNotFound(err.Error()))
		} else {
			stream.Error(res.ErrInternalServer("failed to find logs for container", err))
		}
		return
	}
	defer logs.Close()

	if shouldStream {
		var stop = make(chan struct{})
		socket, err := stream.GetSocketWriter()
		if err != nil {
			stream.Error(res.ErrInternalServer("failed to write to socket", err))
			return
		}
		log.FlushRoutine(socket, logs, stop)
		defer stream.Close()
		defer close(stop)
	} else {
		buf := new(bytes.Buffer)
		buf.ReadFrom(logs)
		render.Render(w, r, res.MsgOK("configured environment variables retrieved",
			"logs", strings.Split(buf.String(), "\n")))
	}
}
