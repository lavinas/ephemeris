package port

type DTOIn interface {
	Validate() error
	GetDomain() Domain
	GetOut(in interface {}) ([]DTOOut, string) 
}

type DTOOut interface {
}
