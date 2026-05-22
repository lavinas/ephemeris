package dto

// ResponseBase represents the base structure for API responses, containing common fields for status and messages.
type ResponseBase struct {
	HttpCode int16
	Status   string
	Message  string
}
