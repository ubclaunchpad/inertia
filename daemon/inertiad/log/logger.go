package log

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

// DaemonLogger is a multilogger used by the daemon to pipe
// output to multiple places depending on context.
type DaemonLogger struct {
	req        *http.Request
	httpWriter http.ResponseWriter
	httpStream bool
	socket     SocketWriter
	io.Writer
}

// LoggerOptions defines configuration for a daemon logger
type LoggerOptions struct {
	Request    *http.Request
	Stdout     io.Writer
	Socket     SocketWriter
	HTTPWriter http.ResponseWriter
	HTTPStream bool
}

// NewLogger creates a new logger
func NewLogger(opts LoggerOptions) *DaemonLogger {
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

	return &DaemonLogger{
		httpWriter: opts.HTTPWriter,
		httpStream: opts.HTTPStream,
		socket:     opts.Socket,
		Writer:     w,
	}
}

// GetSocketWriter retrieves the socketwriter as an io.Writer
func (l *DaemonLogger) GetSocketWriter() (io.Writer, error) {
	if l.socket != nil {
		return NewWebSocketTextWriter(l.socket), nil
	}
	return nil, errors.New("no websocket active")
}

// Println prints to logger's standard writer
func (l *DaemonLogger) Println(a interface{}) {
	fmt.Fprintln(l.Writer, a)
}

// WriteErr directs message and status to http.Error when appropriate
func (l *DaemonLogger) WriteErr(msg string, status int) {
	fmt.Fprintf(l.Writer, "[ERROR %s] %s\n", strconv.Itoa(status), msg)
	if l.socket == nil {
		render.Render(l.httpWriter, l.req, res.Err(l.req, msg, status))
	} else {
		l.Close(CloseOpts{msg, status})
	}
}

// WriteSuccess directs status to Header and sets content type when appropriate
func (l *DaemonLogger) WriteSuccess(msg string, status int) {
	fmt.Fprintf(l.Writer, "[SUCCESS %s] %s\n", strconv.Itoa(status), msg)
	if l.socket == nil && !l.httpStream {
		render.Render(l.httpWriter, l.req, res.Message(l.req, msg, status))
	} else {
		l.Close(CloseOpts{msg, status})
	}
}

// CloseOpts defines options for closing the logger
type CloseOpts struct {
	Message    string
	StatusCode int
}

// Close shuts down the logger
func (l *DaemonLogger) Close(opts ...CloseOpts) error {
	if l.socket != nil && !l.httpStream {
		if opts != nil && len(opts) > 0 {
			return l.socket.CloseHandler()(
				websocket.CloseGoingAway,
				fmt.Sprintf("status %d: %s", opts[0].StatusCode, opts[0].Message))
		}
		return l.socket.CloseHandler()(websocket.CloseGoingAway, "connection closed")
	}
	return nil
}
