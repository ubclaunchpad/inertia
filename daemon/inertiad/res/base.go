package res

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
)

// BaseResponse is the underlying response structure to all responses
type BaseResponse struct {
	HTTPStatusCode int                    `json:"-"`
	Message        string                 `json:"message"`
	RequestID      string                 `json:"request_id"`
	Body           map[string]interface{} `json:"body,omitempty"`
}

func newBaseRequest(r *http.Request, message string, code int, kvs []interface{}) BaseResponse {
	var body = make(map[string]interface{})
	for i := 0; i < len(kvs)-1; i += 2 {
		body[kvs[i].(string)] = kvs[i+1]
	}
	return BaseResponse{
		HTTPStatusCode: code,
		Message:        message,
		RequestID:      middleware.GetReqID(r.Context()),
	}
}
