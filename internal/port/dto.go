package port

type DTO interface {
	GetDomain() Domain
	GetDto(interface{}) (interface{}, string)
	Validate() error
}
