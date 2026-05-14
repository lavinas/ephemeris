package adapter

import (
	"log"
	"os"
	"strings"
)

// SimpleLogger is a simple wrapper around the standard log.Logger
type SimpleLogger struct {
	*log.Logger
	level int
}

// NewSimpleLogger creates a new instance of SimpleLogger
func NewSimpleLogger(output string, level int) (*SimpleLogger, error) {
	if output == "" || output == "stdout" || output == "stderr" {
		return &SimpleLogger{Logger: log.Default(), level: level}, nil
	}
	f, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &SimpleLogger{Logger: log.New(f, "", log.LstdFlags), level: level}, nil
}

// Close closes the logger if it is writing to a file
func (s *SimpleLogger) Close() {
	if s.Logger.Writer() != os.Stdout {
		if f, ok := s.Logger.Writer().(*os.File); ok {
			f.Close()
		}
	}
}

// IPrintf is a identity function for printf, allowing it to be used as a ports.Logger
func (s *SimpleLogger) IPrintf(level int, format string, v ...interface{}) {
	if level > s.level {
		return
	}
	format = strings.Repeat("\t", level) + format
	s.Printf(format, v...)
}
