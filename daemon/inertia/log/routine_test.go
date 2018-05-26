package log

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestFlushRoutine(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// establish socket connection
		var upgrader = websocket.Upgrader{}
		socket, err := upgrader.Upgrade(w, r, nil)
		assert.Nil(t, err)
		defer socket.Close()

		// Everything written to writer will be sent to reader
		reader, writer := io.Pipe()

		// Start continuously reading from reader and writing to socket
		stop := make(chan struct{})
		go FlushRoutine(NewWebSocketTextWriter(socket), reader, stop)

		fmt.Fprintln(writer, "Hello!")
		time.Sleep(time.Millisecond)

		fmt.Fprintln(writer, "Lunch?")
		time.Sleep(time.Millisecond)

		fmt.Fprintln(writer, "Bye!")
		time.Sleep(time.Millisecond)

		// Split sentences should be broadcast to socket as
		// a single message
		fmt.Fprint(writer, "I am")
		time.Sleep(time.Millisecond)
		fmt.Fprint(writer, " awesome")
		time.Sleep(time.Millisecond)
		fmt.Fprintln(writer, " and hungry!!")
		time.Sleep(time.Millisecond)

		// After flusher closed, should not send the next line
		close(stop)
		fmt.Println(writer, "Do I live?")
	}))
	defer testServer.Close()

	// Dial websocket connection
	url, err := url.Parse(testServer.URL)
	assert.Nil(t, err)
	url.Scheme = "ws"
	c, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	assert.Nil(t, err)
	defer c.Close()

	// Read from socket
	i := 0
	for i < 5 {
		_, message, _ := c.ReadMessage()

		switch i {
		case 0:
			assert.Equal(t, "Hello!\n", string(message))
		case 1:
			assert.Equal(t, "Lunch?\n", string(message))
		case 2:
			assert.Equal(t, "Bye!\n", string(message))
		case 3:
			// Three writes received as one message
			assert.Equal(t, "I am awesome and hungry!!\n", string(message))
		case 4:
			assert.Equal(t, "", string(message))
		}

		i++
	}
}
