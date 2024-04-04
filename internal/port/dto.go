package port

type DTOIn interface {
	// Validate is a method that validates the DTOIn
	Validate() error
	// GetDomain is a method that returns the domain of the DTOIn
	GetDomain() []Domain
}

type DTOOut interface {
	GetDTO(domainIn interface{}) interface{}
}
