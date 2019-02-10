package res

import (
	"net/http"

	"github.com/go-chi/render"
)

// MsgResponse is the template for a typical HTTP response for messages
type MsgResponse struct {
	BaseResponse
}

// Render renders a MsgResponse
func (m *MsgResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, m.HTTPStatusCode)
	return nil
}

// Message is a shortcut for non-error statuses
func Message(r *http.Request, message string, code int, kvs ...interface{}) render.Renderer {
	return &MsgResponse{
		BaseResponse: newBaseRequest(r, message, code, kvs),
	}
}
