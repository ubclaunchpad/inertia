package common

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlushRoutine(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		reader, writer := io.Pipe()
		go FlushRoutine(w, reader)

		fmt.Println(writer, "Hello!")
		time.Sleep(time.Millisecond)

		fmt.Println(writer, "Lunch?")
		time.Sleep(time.Millisecond)

		fmt.Println(writer, "Bye!")
		time.Sleep(time.Millisecond)
	}))
	defer testServer.Close()

	resp, err := http.DefaultClient.Get(testServer.URL)
	assert.Nil(t, err)

	reader := bufio.NewReader(resp.Body)
	i := 0
	for i < 3 {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		switch i {
		case 0:
			assert.Equal(t, "Hello!", string(line))
		case 1:
			assert.Equal(t, "Lunch?", string(line))
		case 2:
			assert.Equal(t, "Bye!", string(line))
		}

		i++
	}
}
