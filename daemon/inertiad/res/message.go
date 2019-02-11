package res

// MsgResponse is the template for a typical HTTP response for messages
type MsgResponse struct {
	*baseResponse
}

// Message is a shortcut for non-error statuses
func Message(message string, code int, kvs ...interface{}) *MsgResponse {
	return &MsgResponse{newBaseResponse(message, code, kvs)}
}
