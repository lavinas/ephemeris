package port

type DTOIn interface {
	// Validate is a method that validates the DTOIn
	Validate() error
	// GetDomain is a method that returns the domain of the DTOIn
	GetDomain() Domain
	// GetDomainID is a method that returns the domain just with the ID and other primary fields
	GetDomainID() string
}

type DTOOut interface {
	GetDTO(domainIn interface{}) interface{}
}
