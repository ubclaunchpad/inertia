package log

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

// Streamer is a multilogger used by the daemon to pipe output to multiple
// places depending on context.
type Streamer struct {
	req        *http.Request
	httpWriter http.ResponseWriter
	httpStream bool
	socket     SocketWriter
	io.Writer
}

// StreamerOptions defines configuration for a daemon streamer
type StreamerOptions struct {
	Request    *http.Request
	Stdout     io.Writer
	Socket     SocketWriter
	HTTPWriter http.ResponseWriter
	HTTPStream bool
}

// NewStreamer creates a new streamer. It must be closed if a Socket is provided,
// and one of Error() or Success() should be called.
func NewStreamer(opts StreamerOptions) *Streamer {
	var w io.Writer
	if !opts.HTTPStream {
		// Attempt to create a writer with websocket
		if opts.Socket != nil {
			w = &MultiWriter{
				writers: []io.Writer{opts.Stdout, NewWebSocketTextWriter(opts.Socket)},
			}
		} else {
			w = opts.Stdout
		}
	} else {
		// Attempt to create a writer with HTTPWriter
		w = &MultiWriter{
			writers: []io.Writer{opts.Stdout, opts.HTTPWriter},
		}
	}

	return &Streamer{
		req:        opts.Request,
		httpWriter: opts.HTTPWriter,
		httpStream: opts.HTTPStream,
		socket:     opts.Socket,
		Writer:     w,
	}
}

// GetSocketWriter retrieves the socketwriter as an io.Writer
func (s *Streamer) GetSocketWriter() (io.Writer, error) {
	if s.socket != nil {
		return NewWebSocketTextWriter(s.socket), nil
	}
	return nil, errors.New("no websocket active")
}

// Println prints to logger's standard writer
func (s *Streamer) Println(a ...interface{}) {
	fmt.Fprintln(s.Writer, a...)
}

// Error directs message and status to http.Error when appropriate
func (s *Streamer) Error(res *res.ErrResponse) {
	fmt.Fprintln(s.Writer, res.Error().Error())
	if s.socket == nil {
		render.Render(s.httpWriter, s.req, res)
	} else {
		s.Close(CloseOpts{res.Message, res.HTTPStatusCode})
	}
}

// Success directs status to Header and sets content type when appropriate
func (s *Streamer) Success(res *res.MsgResponse) {
	fmt.Fprintf(s.Writer, "[success %d] %s\n", res.HTTPStatusCode, res.Message)
	if s.socket == nil && !s.httpStream {
		render.Render(s.httpWriter, s.req, res)
	} else {
		s.Close(CloseOpts{res.Message, res.HTTPStatusCode})
	}
}

// CloseOpts defines options for closing the logger
type CloseOpts struct {
	Message    string
	StatusCode int
}

// Close shuts down the logger
func (s *Streamer) Close(opts ...CloseOpts) error {
	if s.socket != nil && !s.httpStream {
		if opts != nil && len(opts) > 0 {
			return s.socket.CloseHandler()(
				websocket.CloseGoingAway,
				fmt.Sprintf("status %d: %s", opts[0].StatusCode, opts[0].Message))
		}
		return s.socket.CloseHandler()(websocket.CloseGoingAway, "connection closed")
	}
	return nil
}
