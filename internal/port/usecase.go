package port

// UseCase is an interface that defines the methods for the use case
type UseCase interface {
	Run(dtoIn interface{}) error
	// Interface is a method that returns the result interface of the use case
	Interface() (interface{}, bool)
	// String is a method that returns the result string of the use case
	String() string
}

// CommandUseCase is an interface that defines the methods for the command use case
// that receives a string command and returns a string response
type CommandUseCase interface {
	// Run is a method that runs the command use case
	Run(string) string
}
