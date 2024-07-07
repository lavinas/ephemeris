package port

// Repository is an interface that defines the methods for the repository
type Repository interface {
	// Migrate is a method that migrates the domain
	// it receives a slice of interfaces that represents the domain
	Migrate(domain []interface{}) error
	// Close is a method that closes the repository
	Close()
	// NewTransaction is a method that creates a new transaction uuid name
	NewTransaction() string 
	// Begin is a method that starts a transaction
	// it receives a string that represents the transaction name
	// if the transaction name is empty, it will be a default transaction
	Begin(tx string) error
	// Commit is a method that commits a transaction
	Commit(tx string) error
	// Rollback is a method that rolls back a transaction
	Rollback(tx string) error
	// Add is a method that adds a new object
	Add(obj interface{}, tx string) error
	// Get is a method that gets an object by its ID
	Get(obj interface{}, id string, tx string) (bool, error)
	// Find is a method that finds an object by its base
	// returns the object, a boolean that indicates if the object was limited and an error
	Find(base interface{}, limit int, tx string, extras ...interface{}) (interface{}, bool, error)
	// Save is a method that saves an object
	Save(obj interface{}, tx string) error
	// Delete is a method that deletes an object by filled fields
	Delete(obj interface{}, tx string, extras ...interface{}) error
}
