package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testBody      = `{"yo":true}`
	testSignature = "sha1=126f2c800419c60137ce748d7672e77b65cf16d6"
	testKey       = "0123456789abcdef"
)

func TestWebhookHandler_notPush(t *testing.T) {
	webhookSecret = testKey
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(webhookHandler)

	handler.ServeHTTP(recorder, getTestWebhookEvent())
	assert.Equal(t, recorder.Code, http.StatusBadRequest)

	b, err := ioutil.ReadAll(recorder.Body)
	assert.Nil(t, err)
	assert.Contains(t, string(b), "Unsupported Github event")
}

func getTestWebhookEvent() *http.Request {
	buf := bytes.NewBufferString(testBody)
	req, err := http.NewRequest("POST", "http://127.0.0.1/webhook", buf)
	if err != nil {
		os.Exit(1)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", "GitHub-Hookshot/539d755")
	req.Header.Set("X-GitHub-Event", "watch")
	req.Header.Set("X-Hub-Signature", testSignature)
	return req
}
