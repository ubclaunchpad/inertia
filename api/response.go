package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

// BaseResponse is the underlying response structure to all responses.
type BaseResponse struct {
	// Basic metadata
	HTTPStatusCode int    `json:"code"`
	RequestID      string `json:"request_id,omitempty"`

	// Message is included in all responses, and is a summary of the server's response
	Message string `json:"message"`

	// Err contains additional context in the event of an error
	Err string `json:"error,omitempty"`

	// Data contains information the server wants to return
	Data interface{} `json:"data,omitempty"`
}

// KV is used for defining specific values to be unmarshalled from BaseResponse
// data
type KV struct {
	Key   string
	Value interface{}
}

// Unmarshal reads the response and unmarshalls the BaseResponse as well any
// requested key-value pairs.
// For example:
//
// 	  var totpResp = &api.TotpResponse{}
//    api.Unmarshal(resp.Body, api.KV{Key: "totp", Value: totpResp})
//
// Values provided in KV.Value MUST be explicit pointers, even if the value is
// a pointer type, ie maps and slices.
func Unmarshal(r io.Reader, kvs ...KV) (*BaseResponse, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read bytes from reader: %s", err.Error())
	}

	// Unmarshal data into a BaseResponse, replacing BaseResponse.Data with a
	// map to preserve raw JSON data in the keys
	var (
		data = make(map[string]json.RawMessage)
		resp = BaseResponse{Data: &data}
	)
	if err := json.Unmarshal(bytes, &resp); err != nil {
		return nil, fmt.Errorf("could not unmarshal data from reader: %s", err.Error())
	}

	// Unmarshal all requested kv-pairs, silently ignoring errors
	for _, kv := range kvs {
		json.Unmarshal(data[kv.Key], kv.Value)
	}

	return &resp, nil
}

// Error returns a summary of an encountered error. For more details, you may
// want to interrogate Data. Returns nil if StatusCode is not an HTTP error
// code, ie if the code is in 1xx, 2xx, or 3xx
func (b *BaseResponse) Error() error {
	if 100 <= b.HTTPStatusCode && b.HTTPStatusCode < 400 {
		return nil
	}
	if b.Err == "" {
		return fmt.Errorf("[error %d] %s", b.HTTPStatusCode, b.Message)
	}
	return fmt.Errorf("[error %d] %s: (%s)", b.HTTPStatusCode, b.Message, b.Err)
}

// TotpResponse is used for sending users their Totp secret and backup codes
type TotpResponse struct {
	TotpSecret  string   `json:"secret"`
	BackupCodes []string `json:"backup_codes"`
}

// DeploymentStatus lists details about the deployed project
type DeploymentStatus struct {
	InertiaVersion       string   `json:"version"`
	Branch               string   `json:"branch"`
	CommitHash           string   `json:"commit_hash"`
	CommitMessage        string   `json:"commit_message"`
	BuildType            string   `json:"build_type"`
	Containers           []string `json:"containers"`
	BuildContainerActive bool     `json:"build_active"`

	// returns tag of latest version on dockerhub
	NewVersionAvailable *string `json:"new_version_available"`
}
