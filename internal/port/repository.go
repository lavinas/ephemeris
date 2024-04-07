package port

// Repository is an interface that defines the methods for the repository
type Repository interface {
	// Migrate is a method that migrates the domain
	Migrate(domain []interface{}) error
	// Begin is a method that starts a transaction
	Begin() error
	// Commit is a method that commits a transaction
	Commit() error
	// Rollback is a method that rolls back a transaction
	Rollback() error
	// Add is a method that adds a new object
	Add(obj interface{}) error
	// Get is a method that gets an object by its ID
	Get(obj interface{}, id string) (bool, error)
	// Find is a method that finds an object by its base
	Find(base interface{}) (interface{}, error)
	// Save is a method that saves an object
	Save(obj interface{}) error
	// Delete is a method that deletes an object by its ID
	Delete(obj interface{}, id string) error
	// Close is a method that closes the repository
	Close()
}
