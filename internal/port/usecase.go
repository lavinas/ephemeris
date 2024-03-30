package port

// UseCase is an interface that defines the methods for the use case
type UseCase interface {
	Command(string) string
	Add(interface{}) (interface{}, string, error)
	Get(interface{}) (interface{}, string, error)
}
