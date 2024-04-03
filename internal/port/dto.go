package port

type DTOIn interface {
	Validate() error
	GetDomain() Domain
}

type DTOOut interface {
	GetDTO(domainIn interface{}) interface{}
}
