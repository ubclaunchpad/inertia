package res

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	"github.com/ubclaunchpad/inertia/api"
)

type baseResponse struct {
	api.BaseResponse
}

func (b *baseResponse) Render(w http.ResponseWriter, r *http.Request) error {
	b.RequestID = reqID(r)
	render.Status(r, b.HTTPStatusCode)
	return nil
}

func newBaseResponse(
	message string,
	code int,
	kvs []interface{},
) *baseResponse {
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
	return &baseResponse{
		api.BaseResponse{
			HTTPStatusCode: code,
			Message:        message,
			Err:            e,
			Data:           data,
		},
	}
}

func reqID(r *http.Request) string {
	if r == nil {
		return ""
	}
	return middleware.GetReqID(r.Context())
}
