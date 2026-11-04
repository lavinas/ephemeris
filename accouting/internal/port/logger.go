package port

// Logger is an interface that defines the logging functionality for the application.
type Logger interface {
	IPrintf(level int, format string, v ...interface{})
	Close()
}

