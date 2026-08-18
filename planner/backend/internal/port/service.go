package port

// Service defines the interface for a service that processes input data and produces output data.
type Service interface {
	Run(in InDTO) (out OutDTO)
}
