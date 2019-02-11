package res

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	"github.com/ubclaunchpad/inertia/api"
)

// BaseResponse is a container class around api.BaseResponse
type BaseResponse struct {
	api.BaseResponse
}

func newBaseResponse(
	message string,
	code int,
	kvs []interface{},
) *BaseResponse {
	e, data := formatData(kvs)
	return &BaseResponse{
		api.BaseResponse{
			HTTPStatusCode: code,
			Message:        message,
			Err:            e,
			Data:           data,
		},
	}
}

// Render implements chi's render.Renderer
func (b *BaseResponse) Render(w http.ResponseWriter, r *http.Request) error {
	b.RequestID = reqID(r)
	render.Status(r, b.HTTPStatusCode)
	return nil
}

func formatData(kvs []interface{}) (e string, data map[string]interface{}) {
	if len(kvs) < 1 {
		return "", nil
	}

	data = make(map[string]interface{})
	var hasNonErrorData = false
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
			hasNonErrorData = true
			data[k] = v
		}
	}

	if !hasNonErrorData {
		return e, nil
	}
	return e, data
}

func reqID(r *http.Request) string {
	if r == nil || r.Context() == nil {
		return ""
	}
	return middleware.GetReqID(r.Context())
}
