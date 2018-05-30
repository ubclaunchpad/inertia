package log

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// DaemonLogger is a multilogger used by the daemon to pipe
// output to multiple places depending on context.
type DaemonLogger struct {
	httpWriter http.ResponseWriter
	socket     SocketWriter
	io.Writer
}

// LoggerOptions defines configuration for a daemon logger
type LoggerOptions struct {
	Stdout     io.Writer
	Socket     SocketWriter
	HTTPWriter http.ResponseWriter
}

// NewLogger creates a new logger
func NewLogger(opts LoggerOptions) *DaemonLogger {
	var w io.Writer
	if opts.Stdout == nil && opts.Socket == nil {
		w = nil
	} else if opts.Stdout != nil && opts.Socket == nil {
		w = opts.Stdout
	} else if opts.Stdout == nil && opts.Socket != nil {
		w = NewWebSocketTextWriter(opts.Socket)
	} else {
		w = io.MultiWriter(opts.Stdout, NewWebSocketTextWriter(opts.Socket))
	}

	return &DaemonLogger{
		httpWriter: opts.HTTPWriter,
		socket:     opts.Socket,
		Writer:     w,
	}
}

// Println prints to logger's standard writer
func (l *DaemonLogger) Println(a interface{}) {
	fmt.Fprintln(l.Writer, a)
}

// WriteErr directs message and status to http.Error when appropriate
func (l *DaemonLogger) WriteErr(msg string, status int) {
	fmt.Fprintf(l.Writer, "[ERROR %s] %s\n", strconv.Itoa(status), msg)
	if l.socket == nil {
		http.Error(l.httpWriter, msg, status)
	}
}

// WriteSuccess directs status to Header and sets content type when appropriate
func (l *DaemonLogger) WriteSuccess(msg string, status int) {
	fmt.Fprintf(l.Writer, "[SUCCESS %s] %s\n", strconv.Itoa(status), msg)
	if l.socket == nil {
		l.httpWriter.Header().Set("Content-Type", "text/html")
		l.httpWriter.WriteHeader(status)
		fmt.Fprintln(l.httpWriter, msg)
	}
}

// Close shuts down the logger
func (l *DaemonLogger) Close() {
	if l.socket != nil {
		l.socket.Close()
	}
}
