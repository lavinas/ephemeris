package port

// Repository is an interface that defines the methods for the repository
type Repository interface {
	Migrate(domain []interface{}) error
	Begin() error
	Commit() error
	Rollback() error
	Add(obj interface{}) error
	Get(obj interface{}, id string) (bool, error)
	Find(base interface{}) (interface{}, error)
	Save(obj interface{}) error
	Delete(obj interface{}, id string) error
	Close()
}
