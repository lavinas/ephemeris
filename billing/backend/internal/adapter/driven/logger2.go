package driven

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultMaxAge       = 7 * 24 * time.Hour
	defaultRotationTime = 24 * time.Hour
)

// Logger2 is an implementation of a logger.
type Logger2 struct {
	zapLogger *zap.Logger
}

// NewLogger2 creates a new instance of Logger2.
func NewLogger2(output string, level int) (*Logger2, error) {
	zapLogger, err := setupZapLogger(output, level)
	if err != nil {
		return nil, err
	}
	return &Logger2{
		zapLogger: zapLogger,
	}, nil
}

// Close flushes any remaining log entries and releases resources.
func (l *Logger2) Close() {
	_ = l.zapLogger.Sync()
}

// Debug logs a debug message.
func (l *Logger2) Debug(msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, fields...)
}

// Info logs an informational message.
func (l *Logger2) Info(msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, fields...)
}

// Error logs an error message.
func (l *Logger2) Error(msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, fields...)
}

// Warn logs a warning message.
func (l *Logger2) Warn(msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, fields...)
}

// IPrintf logs a formatted message (compatível com a interface port.Logger).
func (l *Logger2) IPrintf(level int, format string, v ...interface{}) {
	fmt.Println("Passou")
	msg := fmt.Sprintf(format, v...)
	l.zapLogger.Info(msg)
}

// setupZapLogger sets up and returns a configured zap.Logger instance.
func setupZapLogger(output string, level int) (*zap.Logger, error) {
	var writer io.Writer

	switch output {
	case "", "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		// Se passar "logs/app.log":
		// pattern será: "logs/app-%Y-%m-%d.log"
		ext := filepath.Ext(output)
		base := strings.TrimSuffix(output, ext)
		pattern := base + "-%Y-%m-%d" + ext

		rotator, err := rotatelogs.New(
			pattern,
			rotatelogs.WithLinkName(output),
			rotatelogs.WithMaxAge(defaultMaxAge),
			rotatelogs.WithRotationTime(defaultRotationTime),
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao configurar rotatelogs: %v", err)
		}
		writer = rotator
	}

	writeSyncer := zapcore.AddSync(writer)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		zapcore.Level(level),
	)

	return zap.New(core), nil
}