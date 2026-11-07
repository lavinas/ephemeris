package dto

// ResponseBase represents standard response metadata
type ResponseBase struct {
	HttpCode int    `json:"http_code"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

func NewResponseBase(httpCode int, status, message string) ResponseBase {
	return ResponseBase{
		HttpCode: httpCode,
		Status:   status,
		Message:  message,
	}
}

func (r ResponseBase) GetStatusCode() int {
	return r.HttpCode
}

func (r ResponseBase) GetMessage() string {
	return r.Message
}

