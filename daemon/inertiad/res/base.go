package res

import (
	"net/http"

	"github.com/go-chi/chi/middleware"

	"github.com/ubclaunchpad/inertia/api"
)

func newBaseResponse(r *http.Request, message string, code int, kvs []interface{}) api.BaseResponse {
	var data = make(map[string]interface{})
	var e string
	for i := 0; i < len(kvs)-1; i += 2 {
		var (
			k = kvs[i].(string)
			v = kvs[i+1]
		)
		if k == "error" {
			switch err := v.(type) {
			case error:
				e = err.Error()
			case string:
				e = err
			}
		} else {
			data[k] = v
		}
	}
	return api.BaseResponse{
		HTTPStatusCode: code,
		Message:        message,
		RequestID:      middleware.GetReqID(r.Context()),
		Error:          e,
		Data:           data,
	}
}
