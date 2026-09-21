package port

// Domain represents the interface for domain entities in the system.
type Domain interface {
	// Save persists the domain entity using the provided repository.
	Save(repo Repository) error
	// Find retrieves domain entities based on the specified conditions using the provided repository.
	Find(repo Repository) ([]Domain, error)
}