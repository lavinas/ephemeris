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
    // GetHot gets a object from the database by id in a distinct transaction
	GetHot(obj interface{}, id string) (bool, error)
	// Find is a method that finds an object by its base
	// returns the object, a boolean that indicates if the object was limited and an error
	Find(base interface{}, limit int, extras ...interface{}) (interface{}, bool, error)
	// Save is a method that saves an object
	Save(obj interface{}) error
	// Delete is a method that deletes an object by filled fields
	Delete(obj interface{}, extras ...interface{}) error
	// Close is a method that closes the repository
	Close()
}
