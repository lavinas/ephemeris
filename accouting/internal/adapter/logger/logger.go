package logger

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

// Logger is an implementation of port.Logger using zap and rotatelogs
type Logger struct {
	zapLogger *zap.Logger
}

// NewLogger creates a new instance of Logger
func NewLogger(output string, level int) (*Logger, error) {
	zapLogger, err := setupZapLogger(output, level)
	if err != nil {
		return nil, err
	}
	return &Logger{
		zapLogger: zapLogger,
	}, nil
}

func (l *Logger) Close() {
	_ = l.zapLogger.Sync()
}

func (l *Logger) IPrintf(level int, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	switch level {
	case 1:
		l.zapLogger.Warn(msg)
	case 2:
		l.zapLogger.Error(msg)
	default:
		l.zapLogger.Info(msg)
	}
}

func setupZapLogger(output string, level int) (*zap.Logger, error) {
	var writer io.Writer

	switch output {
	case "", "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
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

