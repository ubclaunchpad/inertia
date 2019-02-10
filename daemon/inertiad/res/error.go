package res

import (
	"net/http"

	"github.com/go-chi/render"
)

// ErrResponse is the template for a typical HTTP response for errors
type ErrResponse struct {
	BaseResponse
}

// Render renders an ErrResponse
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// Err is a basic error response constructor
func Err(r *http.Request, message string, code int, kvs ...interface{}) render.Renderer {
	return &ErrResponse{
		BaseResponse: newBaseRequest(r, message, code, kvs),
	}
}

// ErrInternalServer is a shortcut for internal server errors. It should be
// accompanied by an actual error.
func ErrInternalServer(r *http.Request, message string, err error, kvs ...interface{}) render.Renderer {
	kvs = append(kvs, "error", err.Error())
	return &ErrResponse{
		BaseResponse: newBaseRequest(r, message, http.StatusInternalServerError, kvs),
	}
}

// ErrBadRequest is a shortcut for bad requests
func ErrBadRequest(r *http.Request, message string, kvs ...interface{}) render.Renderer {
	return &ErrResponse{
		BaseResponse: newBaseRequest(r, message, http.StatusBadRequest, kvs),
	}
}

// ErrUnauthorized is a shortcut for unauthorized requests
func ErrUnauthorized(r *http.Request, message string, kvs ...interface{}) render.Renderer {
	return &ErrResponse{
		BaseResponse: newBaseRequest(r, message, http.StatusUnauthorized, kvs),
	}
}

// ErrForbidden is a shortcut for forbidden requests
func ErrForbidden(r *http.Request, message string, kvs ...interface{}) render.Renderer {
	return &ErrResponse{
		BaseResponse: newBaseRequest(r, message, http.StatusForbidden, kvs),
	}
}

// ErrNotFound is a shortcut for forbidden requests
func ErrNotFound(r *http.Request, message string, kvs ...interface{}) render.Renderer {
	return &ErrResponse{
		BaseResponse: newBaseRequest(r, message, http.StatusForbidden, kvs),
	}
}
