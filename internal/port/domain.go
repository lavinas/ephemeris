package port

// Domain is an interface that represents the domain entity
type Domain interface {
	Format(args ...string) error
	GetID() string
}
