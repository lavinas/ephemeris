package port

// Config is an interface that defines the methods for the configuration
type Config interface {
	Get(key string) string
}
