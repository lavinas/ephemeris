package port

// UseCase is an interface that defines the methods for the use case
type UseCase interface {
	Command(string) string
	Add(DTO) (interface{}, string, error)
	Get(DTO) (interface{}, string, error)
}
