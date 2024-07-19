package port

// Domain is an interface that represents the domain entity
type Domain interface {
	// Format is a method that formats the domain entity
	Format(repo Repository, args ...string) error
	// Exists is a method that checks if the domain entity exists
	Load(repo Repository) (bool, error)
	// GetID is a method that returns the id of the domain entity
	GetID() string
	// Get is a method that returns the domain entity
	Get() Domain
	// GetEmpty is a method that returns an empty domain entity with just id
	GetEmpty() Domain

	TableName() string
}
