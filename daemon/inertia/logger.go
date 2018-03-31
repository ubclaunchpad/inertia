package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/ubclaunchpad/inertia/common"
)

// daemonWriter is a custom implementation of io.Writer that
// writes to both os.Stdout and a stream writer if appropriate.
type daemonWriter struct {
	strWriter io.Writer
	stdWriter io.Writer
}

func (d *daemonWriter) Write(p []byte) (n int, err error) {
	if d.strWriter != nil {
		d.strWriter.Write(p)
	}
	return d.stdWriter.Write(p)
}

// daemonLogger is a multilogger used by the daemon to pipe
// output to multiple places depending on context.
type daemonLogger struct {
	stream     bool
	writer     io.Writer
	reader     *io.PipeReader
	httpWriter http.ResponseWriter
}

// newLogger creates a new logger
func newLogger(stream bool, httpWriter http.ResponseWriter) *daemonLogger {
	writer := &daemonWriter{stdWriter: os.Stdout}
	var reader *io.PipeReader
	if stream {
		r, w := io.Pipe()
		go common.FlushRoutine(httpWriter, r)
		writer.strWriter = w
		reader = r
	}
	return &daemonLogger{
		stream:     stream,
		writer:     writer,
		reader:     reader,
		httpWriter: httpWriter,
	}
}

// Println prints to logger's standard writer
func (l *daemonLogger) Println(a interface{}) {
	fmt.Fprintln(l.writer, a)
}

// Err directs message and status to http.Error when appropriate
func (l *daemonLogger) Err(msg string, status int) {
	fmt.Fprintf(l.writer, "[ERROR %s] %s\n", strconv.Itoa(status), msg)
	if !l.stream {
		http.Error(l.httpWriter, msg, status)
	}
}

// Success directs status to Header and sets content type when appropriate
func (l *daemonLogger) Success(msg string, status int) {
	fmt.Fprintf(l.writer, "[SUCCESS %s] %s\n", strconv.Itoa(status), msg)
	if !l.stream {
		l.httpWriter.Header().Set("Content-Type", "text/html")
		l.httpWriter.WriteHeader(status)
		fmt.Fprintln(l.httpWriter, msg)
	}
}

// GetWriter retrieves the logger's stream writer.
func (l *daemonLogger) GetWriter() io.Writer {
	return l.writer
}

// Close shuts down the logger
func (l *daemonLogger) Close() {
	if l.stream {
		l.reader.Close()
	}
}
