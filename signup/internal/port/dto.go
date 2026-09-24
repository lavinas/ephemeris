package port

import "context"

// InDTO represents a generic data transfer object for input of service methods.
type InDTO interface {
	Validate() error
	GetDomain() (Domain, error)
	GetPageParams() (int, int)
	GetOutDTO(httpCode int, status, message string, data interface{}) OutDTO
}

// OutDTO represents a generic data transfer object for output of service methods.
type OutDTO interface {
	GetStatusCode() int
}

// EventEmittable represents an input DTO capable of emitting events via a publisher.
type EventEmittable interface {
	EmitEvent(ctx context.Context, publisher CustomerEventPublisher) error
}
