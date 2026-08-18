package port

// Repository defines the interface for interacting with the data storage layer for invoices.
type Repository interface {
	// Save saves a new invoice to the database and returns the created invoice with its ID.
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
	Save(model interface{}) error
	Find(model interface{}, page int, pagesize int, conditions ...interface{}) error
	Close() error
}
