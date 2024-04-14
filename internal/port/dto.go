package port

type DTOIn interface {
	// Validate is a method that validates the DTOIn
	Validate(repo Repository) error
	// GetDomain is a method that returns the domain of the DTOIn
	GetDomain() []Domain
	// GetOut is a method that returns the DTOOut
	GetOut() DTOOut
}

type DTOOut interface {
	// GetDTO is a method that returns the DTOOut
	GetDTO(domainIn interface{}) []DTOOut
}
