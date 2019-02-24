package res

import "net/http"

// MsgResponse is the template for a typical HTTP response for messages
type MsgResponse struct {
	*BaseResponse
}

// Msg is a shortcut for non-error statuses
func Msg(message string, code int, kvs ...interface{}) *MsgResponse {
	return &MsgResponse{newBaseResponse(message, code, kvs)}
}

// MsgOK is a shortcut for an ok-status response
func MsgOK(message string, kvs ...interface{}) *MsgResponse {
	return &MsgResponse{newBaseResponse(message, http.StatusOK, kvs)}
}
