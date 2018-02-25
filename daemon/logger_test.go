package daemon

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDaemonWriter(t *testing.T) {
	var b1 bytes.Buffer
	var b2 bytes.Buffer
	writer := daemonWriter{
		strWriter: &b1,
		stdWriter: &b2,
	}
	writer.Write([]byte("whoah!"))
	assert.Equal(t, "whoah!", b1.String())
	assert.Equal(t, "whoah!", b2.String())
}

func TestNewLogger(t *testing.T) {
	w := httptest.NewRecorder()
	logger := newLogger(true, w)
	_, ok := logger.GetWriter().(*daemonWriter)
	assert.True(t, ok)
}

func TestPrintln(t *testing.T) {
	var b bytes.Buffer
	logger := &daemonLogger{writer: &b}
	logger.Println("what???")
	assert.Equal(t, "what???\n", b.String())
}

func TestErr(t *testing.T) {
	var b bytes.Buffer
	w := httptest.NewRecorder()

	// Test streaming
	logger := &daemonLogger{
		stream:     true,
		writer:     &b,
		httpWriter: w,
	}
	logger.Err("Wee!", 200)
	assert.Equal(t, "[ERROR 200] Wee!", b.String())

	// Test direct to httpResponse
	logger.stream = false
	logger.Err("Wee!", 200)
	body, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Wee!\n", string(body))
	assert.Equal(t, 200, w.Code)
}

func TestSuccess(t *testing.T) {
	var b bytes.Buffer
	w := httptest.NewRecorder()

	// Test streaming
	logger := &daemonLogger{
		stream:     true,
		writer:     &b,
		httpWriter: w,
	}
	logger.Success("Wee!", 200)
	assert.Equal(t, "[SUCCESS 200] Wee!", b.String())

	// Test direct to httpResponse
	logger.stream = false
	logger.Success("Wee!", 200)
	body, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Wee!\n", string(body))
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
}
