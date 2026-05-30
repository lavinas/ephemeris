package dto

// ResponseBase represents the base structure for API responses, containing common fields for status and messages.
type ResponseBase struct {
	HttpCode int16
	Status   string
	Message  string
}

// NewResponseBase creates a new instance of ResponseBase with the provided HTTP code, status, and message.
func NewResponseBase(httpCode int16, status, message string) ResponseBase {
	return ResponseBase{
		HttpCode: httpCode,
		Status:   status,
		Message:  message,
	}
}

// GetHTTPCode returns the HTTP status code of the response.
func (r ResponseBase) GetStatusCode() int16 {
	return r.HttpCode
}
