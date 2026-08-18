package port

// InDTO represents a generic data transfer object for input of service methods.
type InDTO interface {
	Validate(repo Repository) error
	Reset()
}

// OutDTO represents a generic data transfer object for output of service methods.
type OutDTO interface {
	GetStatusCode() int
}
