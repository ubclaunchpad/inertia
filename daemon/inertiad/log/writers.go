package log

import (
	"io"

	"github.com/gorilla/websocket"
)

// SocketWriter is an interface for writing to websocket connections
type SocketWriter interface {
	WriteMessage(messageType int, bytes []byte) error
	Close() error
}

// WebSocketWriter wraps a SocketWriter in an io.Writer
type WebSocketWriter struct {
	messageType  int
	socketWriter SocketWriter
}

func (d *WebSocketWriter) Write(p []byte) (int, error) {
	return len(p), d.socketWriter.WriteMessage(d.messageType, p)
}

// NewWebSocketTextWriter returns an io.Writer version of SocketWriter
func NewWebSocketTextWriter(socket SocketWriter) *WebSocketWriter {
	if socket == nil {
		return nil
	}
	return &WebSocketWriter{
		messageType:  websocket.TextMessage,
		socketWriter: socket,
	}
}

// TwoWriter writes to two writers without caring whether one fails
type TwoWriter struct {
	writer1 io.Writer
	writer2 io.Writer
}

func (t *TwoWriter) Write(p []byte) (int, error) {
	var (
		len1 int
		len2 int
		err1 error
		err2 error
	)
	if t.writer1 != nil {
		len1, err1 = t.writer1.Write(p)
	}
	if t.writer2 != nil {
		len2, err2 = t.writer2.Write(p)
	}
	if err1 != nil {
		return len1, err1
	}
	return len2, err2
}
