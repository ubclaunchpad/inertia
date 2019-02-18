package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_allowedRequest(t *testing.T) {
	type args struct {
		r       *http.Request
		methods []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"no methods provided", args{&http.Request{Method: "POST"}, []string{}}, true},
		{"request allowed", args{&http.Request{Method: "POST"}, []string{"POST"}}, true},
		{"request disallowed", args{&http.Request{Method: "GET"}, []string{"POST"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allowedRequest(tt.args.r, tt.args.methods...); got != tt.want {
				t.Errorf("allowedRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithMethods(t *testing.T) {
	type args struct {
		requestMethod  string
		allowedMethods []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"valid request", args{"GET", []string{"GET"}}, true},
		{"invalid request", args{"POST", []string{"GET"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				called   = false
				recorder = httptest.NewRecorder()
				handler  = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true
				})
			)
			req, err := http.NewRequest(tt.args.requestMethod, "/down", nil)
			assert.NoError(t, err)

			WithMethods(handler, tt.args.allowedMethods...).
				ServeHTTP(recorder, req)

			assert.Equal(t, tt.want, called)
		})
	}
}
