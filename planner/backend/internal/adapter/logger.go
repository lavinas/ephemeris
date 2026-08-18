package adapter

import (
	"log"
	"os"
	"strings"
)

// Logger is a simple wrapper around the standard log.Logger
type Logger struct {
	*log.Logger
	output *string
	level  int
}

// NewLogger creates a new instance of Logger
func NewLogger(output string, level int) (*Logger, error) {
	if output == "" || output == "stdout" || output == "stderr" {
		return &Logger{Logger: log.Default(), output: nil, level: level}, nil
	}
	f, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{Logger: log.New(f, "", log.LstdFlags), output: &output, level: level}, nil
}

// Close closes the logger if it is writing to a file
func (s *Logger) Close() {
	if s.Logger.Writer() != os.Stdout {
		if f, ok := s.Logger.Writer().(*os.File); ok {
			f.Close()
		}
	}
}

// IPrintf is a identity function for printf, allowing it to be used as a ports.Logger
func (s *Logger) IPrintf(level int, format string, v ...interface{}) {
	s.testAndReopen()
	if s.level != 0 && level > s.level {
		return
	}
	format = strings.Repeat("\t", level) + format
	s.Printf(format, v...)
}

// testAndReopen checks if the log file exists and reopens it if necessary
func (s *Logger) testAndReopen() {
	if s.output != nil {
		if _, err := os.Stat(*s.output); os.IsNotExist(err) {
			// If the file does not exist, create it
			f, err := os.OpenFile(*s.output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Printf("Error creating log file: %v", err)
				return
			}
			s.SetOutput(f)
		}
	}
}
