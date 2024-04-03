package port

// UseCase is an interface that defines the methods for the use case
type UseCase interface {
	Command(string) string
	Add(in DTOIn) ([]DTOOut, string, error)
	Get(in DTOIn) ([]DTOOut, string, error)
	Up(in DTOIn) ([]DTOOut, string, error)
}
