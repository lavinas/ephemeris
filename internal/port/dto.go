package port

// DTO is an interface that defines the methods for all dtos
type DTO interface {
	GetObject() string
	GetAction() string
}
