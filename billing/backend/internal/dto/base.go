package dto

import (
	"fmt"
	"github.com/nyaruka/phonenumbers"
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


// ValidatePhoneNumber checks if the provided phone number is valid and formats it to E.164 standard.
func ValidateCellNumber(phone string) (string, error) {
	num, err := phonenumbers.Parse(phone, "BR")
	if err != nil {
		return "", fmt.Errorf("invalid phone number format")
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("invalid phone number")
	}
	ptype := phonenumbers.GetNumberType(num)
	if ptype != phonenumbers.MOBILE && ptype != phonenumbers.UNKNOWN && ptype != phonenumbers.FIXED_LINE_OR_MOBILE {
		return "", fmt.Errorf("phone number is not a mobile number")
	}
	return phonenumbers.Format(num, phonenumbers.E164), nil
}