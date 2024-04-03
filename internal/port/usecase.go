package port

// UseCase is an interface that defines the methods for the use case
type UseCase interface {
	SetRepo(repo Repository)
	SetLog(log Logger)
	Run(dtoIn interface{}) error
	Interface() interface{}
	String() string
}

// CommandUseCase is an interface that defines the methods for the command use case
// that receives a string command and returns a string response
type CommandUseCase interface {
	Run(string) string
}
