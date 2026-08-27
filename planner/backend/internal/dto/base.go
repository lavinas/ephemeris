package dto

var (
	validStatuses = map[string]bool{
		"realizada":            true,
		"cancelada_cobrar":     true,
		"cancelada_nao_cobrar": true,
	}
	validServices = map[string]bool{
		"aula/canto": true,
		"aula/piano":    true,
	}
)

// ResponseBase represents the base structure for API responses, containing common fields for status and messages.
type ResponseBase struct {
	HttpCode int    `json:"http_code"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// NewResponseBase creates a new instance of ResponseBase with the provided HTTP code, status, and message.
func NewResponseBase(httpCode int, status, message string) ResponseBase {
	return ResponseBase{
		HttpCode: httpCode,
		Status:   status,
		Message:  message,
	}
}

// GetHTTPCode returns the HTTP status code of the response.
func (r ResponseBase) GetStatusCode() int {
	return r.HttpCode
}
