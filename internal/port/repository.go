package port

// Repository is an interface that defines the methods for the repository
type Repository interface {
	Add(obj interface{}) error
	Get(obj interface{}, id string) (bool, error)
	Find(obj interface{}) (bool, error)
	Delete(obj interface{}, id string) error
	Close()
}
