package dto

type PingResponse struct {
	ResponseBase
}

// NewPingResponse creates a new instance of PingResponse with the provided HTTP code, status, and message.
func NewPingResponse(httpCode int, status, message string) PingResponse {
	return PingResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}
