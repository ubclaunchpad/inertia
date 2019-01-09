package daemon

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/cfg"
)

const (
	testBody      = `{"yo":true}`
	testSignature = "sha1=126f2c800419c60137ce748d7672e77b65cf16d6"
	testKey       = "0123456789abcdef"
)

func Test_webhookHandler(t *testing.T) {
	type args struct {
		secret  string
		headers map[string]string
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantErr  string
	}{
		{"okay but unsupported", args{
			testKey,
			map[string]string{
				"content-type":    "application/json",
				"User-Agent":      "GitHub-Hookshot/539d755",
				"X-GitHub-Event":  "watch",
				"X-Hub-Signature": testSignature,
			},
		}, http.StatusBadRequest, "unsupported Github event"},
		{"no signature", args{
			testKey,
			map[string]string{
				"content-type":   "application/json",
				"User-Agent":     "GitHub-Hookshot/539d755",
				"X-GitHub-Event": "push",
			},
		}, http.StatusBadRequest, "missing signature"},
		{"no secret", args{
			"",
			map[string]string{
				"content-type":    "application/json",
				"User-Agent":      "GitHub-Hookshot/539d755",
				"X-GitHub-Event":  "push",
				"X-Hub-Signature": testSignature,
			},
		}, http.StatusBadRequest, "payload signature check failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s = &Server{
				state: cfg.Config{WebhookSecret: tt.args.secret},
			}
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(s.webhookHandler)

			handler.ServeHTTP(recorder, getTestWebhookEvent(tt.args.headers))
			assert.Equal(t, recorder.Code, tt.wantCode)

			b, err := ioutil.ReadAll(recorder.Body)
			assert.Nil(t, err)
			assert.Contains(t, string(b), tt.wantErr)
		})
	}
}

func getTestWebhookEvent(headers map[string]string) *http.Request {
	buf := bytes.NewBufferString(testBody)
	req, err := http.NewRequest("POST", "http://127.0.0.1/webhook", buf)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}
