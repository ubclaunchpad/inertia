package log

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockSocketWriter struct {
	bytes.Buffer
}

func (m *mockSocketWriter) CloseHandler() func(code int, text string) error {
	return func(code int, text string) error { return nil }
}
func (m *mockSocketWriter) WriteMessage(t int, bytes []byte) error {
	_, err := m.Buffer.Write(bytes)
	return err
}
func (m *mockSocketWriter) getWrittenBytes() *bytes.Buffer { return &m.Buffer }

func TestNewLogger(t *testing.T) {
	logger := NewLogger(LoggerOptions{})
	assert.NotNil(t, logger)
}

func TestWrite(t *testing.T) {
	var b bytes.Buffer
	writer := NewLogger(LoggerOptions{Stdout: &b})
	writer.Write([]byte("whoah!"))
	assert.Equal(t, "whoah!", b.String())
}

func TestWriteMulti(t *testing.T) {
	var b1 bytes.Buffer
	socketWriter := &mockSocketWriter{}
	writer := NewLogger(LoggerOptions{Stdout: &b1, Socket: socketWriter})
	writer.Write([]byte("whoah!"))
	assert.Equal(t, "whoah!", b1.String())
	assert.Equal(t, "whoah!", socketWriter.getWrittenBytes().String())
}

func TestPrintln(t *testing.T) {
	var b bytes.Buffer
	logger := &DaemonLogger{Writer: &b}
	logger.Println("what???")
	assert.Equal(t, "what???\n", b.String())
}

func TestErr(t *testing.T) {
	var b bytes.Buffer
	w := httptest.NewRecorder()

	// Test streaming
	logger := &DaemonLogger{
		Writer:     &b,
		httpWriter: w,
		socket:     &mockSocketWriter{},
	}
	logger.WriteErr("Wee!", 200)
	assert.Equal(t, "[ERROR 200] Wee!\n", b.String())

	// Test direct to httpResponse
	logger.socket = nil
	logger.WriteErr("Wee!", 200)
	body, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Wee!\n", string(body))
	assert.Equal(t, 200, w.Code)
}

func TestSuccess(t *testing.T) {
	var b bytes.Buffer
	w := httptest.NewRecorder()

	// Test streaming
	logger := &DaemonLogger{
		httpWriter: w,
		Writer:     &b,
		socket:     &mockSocketWriter{},
	}
	logger.WriteSuccess("Wee!", 200)
	assert.Equal(t, "[SUCCESS 200] Wee!\n", b.String())

	// Test direct to httpResponse
	logger.socket = nil
	logger.WriteSuccess("Wee!", 200)
	body, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Wee!\n", string(body))
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
}
