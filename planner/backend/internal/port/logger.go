package port

// Logger is an interface that defines the logging functionality for the application.
type Logger interface {
	// IPrintf logs a formatted message at the specified log level.
	// The log level can be used to control the verbosity of the logging output.
	IPrintf(level int, format string, v ...interface{})
	// Close closes the logger and releases any resources it holds.
	Close()
}
