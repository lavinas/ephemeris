package port

// UseCase is an interface that defines the methods for the use case
type UseCase interface {
	Command(string) string
	ClientAdd(DTO) (string, error)
	ClientGet(DTO) (string, error)
}
