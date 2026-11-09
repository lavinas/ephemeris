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

// Logger gerencia logs estruturados com zap e rotação diária de arquivos
type Logger struct {
	zapLogger *zap.Logger
	sugar     *zap.SugaredLogger
}

// NewLogger cria uma nova instância de Logger salvando em arquivo rotativo e stdout
func NewLogger(output string, level int) (*Logger, error) {
	zapLogger, err := setupZapLogger(output, level)
	if err != nil {
		return nil, err
	}
	return &Logger{
		zapLogger: zapLogger,
		sugar:     zapLogger.Sugar(),
	}, nil
}

// Close sincroniza e descarrega os buffers de log
func (l *Logger) Close() {
	_ = l.zapLogger.Sync()
}

// Info registra mensagem informativa
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, fields...)
}

// Infof registra mensagem informativa formatada
func (l *Logger) Infof(format string, args ...any) {
	l.sugar.Infof(format, args...)
}

// Warn registra mensagem de aviso
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, fields...)
}

// Warnf registra mensagem de aviso formatada
func (l *Logger) Warnf(format string, args ...any) {
	l.sugar.Warnf(format, args...)
}

// Error registra mensagem de erro
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, fields...)
}

// Errorf registra mensagem de erro formatada
func (l *Logger) Errorf(format string, args ...any) {
	l.sugar.Errorf(format, args...)
}

// Debug registra mensagem de depuração
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, fields...)
}

// Debugf registra mensagem de depuração formatada
func (l *Logger) Debugf(format string, args ...any) {
	l.sugar.Debugf(format, args...)
}

// LogAPICall registra de forma estruturada as chamadas feitas para as APIs do Pluggy
func (l *Logger) LogAPICall(method, endpoint string, statusCode int, duration time.Duration, itemsCount int, err error) {
	fields := []zap.Field{
		zap.String("system", "pluggy-sync"),
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.Int("status_code", statusCode),
		zap.Int64("duration_ms", duration.Milliseconds()),
		zap.Int("items_returned", itemsCount),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.zapLogger.Error(fmt.Sprintf("[Pluggy API] %s %s falhou (status %d) em %dms: %v", method, endpoint, statusCode, duration.Milliseconds(), err), fields...)
	} else {
		l.zapLogger.Info(fmt.Sprintf("[Pluggy API] %s %s -> %d (%dms, %d itens)", method, endpoint, statusCode, duration.Milliseconds(), itemsCount), fields...)
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
		dir := filepath.Dir(output)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0777); err != nil {
				return nil, fmt.Errorf("erro ao criar diretório de log %s: %w", dir, err)
			}
		}

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
			return nil, fmt.Errorf("erro ao configurar rotatelogs: %w", err)
		}
		// Escreve tanto no arquivo rotativo quanto no console (stdout)
		writer = io.MultiWriter(os.Stdout, rotator)
	}

	writeSyncer := zapcore.AddSync(writer)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		zapcore.Level(level),
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)), nil
}

