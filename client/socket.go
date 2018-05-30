package client

// SocketReader is an interface to a websocket connection
type SocketReader interface {
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}
