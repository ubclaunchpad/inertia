package log

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"

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

func TestNewStreamer(t *testing.T) {
	logger := NewStreamer(StreamerOptions{})
	assert.NotNil(t, logger)
}

func TestWrite(t *testing.T) {
	var b bytes.Buffer
	writer := NewStreamer(StreamerOptions{Stdout: &b})
	writer.Write([]byte("whoah!"))
	assert.Equal(t, "whoah!", b.String())
}

func TestWriteMulti(t *testing.T) {
	var b1 bytes.Buffer
	socketWriter := &mockSocketWriter{}
	writer := NewStreamer(StreamerOptions{Stdout: &b1, Socket: socketWriter})
	writer.Write([]byte("whoah!"))
	assert.Equal(t, "whoah!", b1.String())
	assert.Equal(t, "whoah!", socketWriter.getWrittenBytes().String())
}

func TestPrintln(t *testing.T) {
	var b bytes.Buffer
	logger := &Streamer{Writer: &b}
	logger.Println("what???")
	assert.Equal(t, "what???\n", b.String())
}

func TestErr(t *testing.T) {
	var b bytes.Buffer
	w := httptest.NewRecorder()

	// Test streaming
	var req = httptest.NewRequest("GET", "/asdf", nil)
	logger := &Streamer{
		req:        req,
		Writer:     &b,
		httpWriter: w,
		socket:     &mockSocketWriter{},
	}
	logger.Error(res.ErrBadRequest("Wee!"))
	assert.Equal(t, "[error 400] Wee!\n", b.String())

	// Test direct to httpResponse
	logger.socket = nil
	logger.Error(res.ErrBadRequest("Wee!"))
	body, err := api.Unmarshal(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Wee!", body.Message)
	assert.Equal(t, 400, w.Code)
}

func TestSuccess(t *testing.T) {
	var b bytes.Buffer
	w := httptest.NewRecorder()

	// Test streaming
	var req = httptest.NewRequest("GET", "/asdf", nil)
	logger := &Streamer{
		req:        req,
		httpWriter: w,
		Writer:     &b,
		socket:     &mockSocketWriter{},
	}
	logger.Success(res.MsgOK("Wee!"))
	assert.Equal(t, "[success 200] Wee!\n", b.String())

	// Test direct to httpResponse
	logger.socket = nil
	logger.Success(res.MsgOK("Wee!"))
	body, err := api.Unmarshal(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Wee!", body.Message)
	assert.Equal(t, 200, w.Code)
}
