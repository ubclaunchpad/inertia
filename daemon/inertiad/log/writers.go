package log

import (
	"io"
	"net/http"

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

func (w *WebSocketWriter) Write(p []byte) (int, error) {
	return len(p), w.socketWriter.WriteMessage(w.messageType, p)
}

// Close closes the socket writer's websocket.
func (w *WebSocketWriter) Close() error {
	return w.socketWriter.Close()
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

// MultiWriter writes to list of writers without caring whether one fails, and
// flushes if writer is flushable
type MultiWriter struct {
	writers []io.Writer
}

func (m *MultiWriter) Write(p []byte) (int, error) {
	var (
		lastLen int
		lastErr error
	)
	for i := 0; i < len(m.writers); i++ {
		writer := m.writers[i]
		if writer == nil {
			continue
		}
		len, err := writer.Write(p)
		if err != nil {
			lastErr = err
		} else {
			lastLen = len
			if f, ok := writer.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
	return lastLen, lastErr
}
