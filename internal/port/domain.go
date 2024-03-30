package port

// Domain is an interface that represents the domain entity
type Domain interface {
	Validate() error
	Format()
	String() string
	GetID() string
}
