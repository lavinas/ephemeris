package port

// Logger is an interface that defines the methods for logging
type Logger interface {
	Print(v ...any)
	Printf(format string, v ...any)
	Println(v ...any)
}
