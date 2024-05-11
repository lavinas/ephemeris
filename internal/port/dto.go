package port

type DTOIn interface {
	// Validate is a method that validates the DTOIn
	Validate(repo Repository) error
	// GetCommand is a method that returns the command of the DTOIn
	GetCommand() string
	// GetDomain is a method that returns the domain of the DTOIn
	GetDomain() []Domain
	// GetOut is a method that returns the DTOOut
	GetOut() DTOOut
	// GetInstructions is a method that returns the instructions of the DTOIn for a given domain
	GetInstructions(domain Domain) (Domain, []interface{}, error)
}

type DTOOut interface {
	// GetDTO is a method that returns the DTOOut
	GetDTO(domainIn interface{}) []DTOOut
}
