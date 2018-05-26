package log

import "github.com/gorilla/websocket"

// SocketWriter is an interface for writing to websocket connections
type SocketWriter interface {
	WriteMessage(messageType int, bytes []byte) error
	Close() error
}

// WebSocketWriter waps a SocketWriter in an io.Writer
type WebSocketWriter struct {
	messageType  int
	socketWriter SocketWriter
}

func (d *WebSocketWriter) Write(p []byte) (int, error) {
	err := d.socketWriter.WriteMessage(d.messageType, p)
	return len(p), err
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
