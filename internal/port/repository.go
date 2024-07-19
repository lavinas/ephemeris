package port

// Repository is an interface that defines the methods for the repository
type Repository interface {
	// Migrate is a method that migrates the domain
	// it receives a slice of interfaces that represents the domain
	Migrate(domain []interface{}) error
	// Close is a method that closes the repository
	Close()
	// Begin is a method that starts a transaction
	// it receives a string that represents the transaction name
	// if the transaction name is empty, it will be a default transaction
	Begin() interface{}
	// Commit is a method that commits a transaction
	Commit(tx interface{}) error
	// Rollback is a method that rolls back a transaction
	Rollback(tx interface{}) error
	// Add is a method that adds a new object
	Add(tx interface{}, obj interface{}) error
	// Get is a method that gets an object by its ID
	Get(tx interface{}, obj interface{}, id string, lock bool) (bool, error)
	// Find is a method that finds an object by its base
	// returns the object, a boolean that indicates if the object was limited and an error
	Find(tx interface{}, obj interface{}, limit int, lock bool, extras ...interface{}) (interface{}, bool, error)
	// Save is a method that saves an object
	Save(tx interface{}, obj interface{}) error
	// Delete is a method that deletes an object by filled fields
	Delete(tx interface{}, obj interface{}, extras ...interface{}) error
}
