package port

// Domain represents the interface for domain entities in the system.
type Domain interface {
	// Validate validates the domain entity.
	Validate() error
	// Save persists the domain entity using the provided repository.
	Save() error
	// Find retrieves domain entities based on the specified conditions using the provided repository.
	Find(page, pagesize int) ([]Domain, error)
}
