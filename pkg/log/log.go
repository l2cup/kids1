package log

import (
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DebugVerbosity = "DEBUG"
	InfoVerbosity  = "INFO"
	WarnVerbosity  = "WARN"
	ErrorVerbosity = "ERROR"
)

type Logger struct {
	zlog *zap.SugaredLogger
}

type Config struct {
	LogVerbosity string
}

func NewLogger(config *Config) (*Logger, error) {

	zapLogger, err := newZapLogger(config)

	if err != nil {
		return nil, errors.Wrap(err, "new zap logger")
	}

	zap.ReplaceGlobals(zapLogger)

	return &Logger{
		zlog: zapLogger.Sugar(),
	}, nil
}

func newZapLogger(config *Config) (*zap.Logger, error) {

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.LineEnding = zapcore.DefaultLineEnding

	jsonEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	atom := zap.NewAtomicLevel()
	atom.SetLevel(mapVerbosityLevel(config.LogVerbosity))

	core := zapcore.NewCore(jsonEncoder, zapcore.Lock(os.Stdout), atom)

	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	logger := zap.New(core).WithOptions(options...)

	defer logger.Sync()

	return logger, nil
}

func (log *Logger) Debug(msg string, fields ...interface{}) {
	log.zlog.Debugw(msg, fields...)
}

func (log *Logger) Info(msg string, fields ...interface{}) {
	log.zlog.Infow(msg, fields...)
}

func (log *Logger) Warn(msg string, fields ...interface{}) {
	log.zlog.Warnw(msg, fields...)
}

func (log *Logger) Error(msg string, fields ...interface{}) {
	log.zlog.Errorw(msg, fields...)
}

func (log *Logger) Fatal(msg string, fields ...interface{}) {
	log.zlog.Fatalw(msg, fields...)
}

func mapVerbosityLevel(verbosity string) zapcore.Level {
	switch verbosity {
	case DebugVerbosity:
		return zapcore.DebugLevel
	case InfoVerbosity:
		return zapcore.InfoLevel
	case WarnVerbosity:
		return zapcore.WarnLevel
	case ErrorVerbosity:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
