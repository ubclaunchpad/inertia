package log

import "github.com/gorilla/websocket"

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
