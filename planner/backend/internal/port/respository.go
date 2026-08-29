package port

// Repository defines the interface for interacting with the data storage layer for invoices.
type Repository interface {
	// Save saves a new invoice to the database and returns the created invoice with its ID.
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
	Save(model interface{}) error
	Find(page, pagesize int, conditions map[string]interface{}) ([]interface{}, error)
	FindCount(conditions map[string]interface{}) (int64, error)
	FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error)
	Close() error
}
