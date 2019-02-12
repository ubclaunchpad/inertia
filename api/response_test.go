package api

import (
	"bytes"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantErr  bool
	}{
		{"invalid data",
			args{nil},
			0,
			true},
		{"invalid json",
			args{
				[]byte(`{
				"code":200,
				"request_id":"bobbook/2Mch7LMzhj-000001",
				"mess`),
			},
			0,
			true},
		{"ok",
			args{
				[]byte(`{
					"code":200,
					"request_id":"bobbook/2Mch7LMzhj-000001",
					"message":"session created",
					"data":{
						"token":"blah"
					}
				}`),
			},
			200,
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotToken string
			var gotKV = []KV{{Key: "token", Value: &gotToken}}
			got, err := Unmarshal(bytes.NewReader(tt.args.bytes), gotKV...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.wantCode != got.HTTPStatusCode {
				t.Errorf("Unmarshal() code = %v, want %v", got.HTTPStatusCode, tt.wantCode)
			}
			for _, kv := range gotKV {
				if kv.Value == "" {
					t.Error("Unmarshal() kv is empty")
				}
			}
		})
	}
}

func TestBaseResponse_Error(t *testing.T) {
	type fields struct {
		HTTPStatusCode int
		Message        string
		Err            string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"not an error",
			fields{200, "hi", ""},
			false},
		{"error with only message",
			fields{400, "hi", ""},
			true},
		{"error with message and error context",
			fields{400, "hi", "oh no"},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseResponse{
				HTTPStatusCode: tt.fields.HTTPStatusCode,
				Message:        tt.fields.Message,
				Err:            tt.fields.Err,
			}
			if err := b.Error(); (err != nil) != tt.wantErr {
				t.Errorf("BaseResponse.Error() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
