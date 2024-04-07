package port

// Config is an interface that defines the methods for the configuration
type Config interface {
	// Get is a method that returns the value of the key
	Get(key string) string
}
