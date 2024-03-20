package port

// UseCase is an interface that defines the methods for the use case
type UseCase interface {
	AddClient(id, name, nickname, email, phone, contact, document string) error
	// GetClient(id string) (*Client, error)
}