package port

// HTTPHandler defines the interface for an HTTP server handler.
type HTTPHandler interface {
	Run(addr string) error
}