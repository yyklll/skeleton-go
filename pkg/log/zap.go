package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var DefaultZapLogLevel = "info"

func InitZapConfig(lvl string) zap.Config {
	return zap.Config{
		Level: zap.NewAtomicLevelAt(ConvertToZapLevel(lvl)),

		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},

		Encoding: "json",

		// copied from "zap.NewProductionEncoderConfig" with some updates
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},

		// Use "/dev/null" to discard all
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// ConvertToZapLevel converts log level string to zapcore.Level.
func ConvertToZapLevel(lvl string) zapcore.Level {
	switch lvl {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		panic(fmt.Sprintf("unknown level %q", lvl))
	}
}

func ConvertZapLevelToString(lvl zapcore.Level) string {
	switch lvl {
	case zap.DebugLevel:
		return "debug"
	case zap.InfoLevel:
		return "info"
	case zap.WarnLevel:
		return "warn"
	case zap.ErrorLevel:
		return "error"
	case zap.DPanicLevel:
		return "dpanic"
	case zap.PanicLevel:
		return "panic"
	case zap.FatalLevel:
		return "fatal"
	default:
		panic(fmt.Sprintf("unknown level %q", lvl))
	}
}

func NewZapLogger(lcfg zap.Config) (Logger, error) {
	lg, err := lcfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &zapLogger{lg: lg, sugar: lg.Sugar(), cfg: lcfg}, nil
}

type zapLogger struct {
	lg    *zap.Logger
	sugar *zap.SugaredLogger
	cfg   zap.Config
}

func (zl *zapLogger) Debug(args ...interface{}) {
	if !zl.lg.Core().Enabled(zapcore.DebugLevel) {
		return
	}
	zl.sugar.Debug(args...)
}

func (zl *zapLogger) Debugln(args ...interface{}) {
	if !zl.lg.Core().Enabled(zapcore.DebugLevel) {
		return
	}
	zl.sugar.Debug(args...)
}

func (zl *zapLogger) Debugf(format string, args ...interface{}) {
	if !zl.lg.Core().Enabled(zapcore.DebugLevel) {
		return
	}
	zl.sugar.Debugf(format, args...)
}

func (zl *zapLogger) Info(args ...interface{}) {
	if !zl.lg.Core().Enabled(zapcore.DebugLevel) {
		return
	}
	zl.sugar.Info(args...)
}

func (zl *zapLogger) Infoln(args ...interface{}) {
	if !zl.lg.Core().Enabled(zapcore.DebugLevel) {
		return
	}
	zl.sugar.Info(args...)
}

func (zl *zapLogger) Infof(format string, args ...interface{}) {
	if !zl.lg.Core().Enabled(zapcore.DebugLevel) {
		return
	}
	zl.sugar.Infof(format, args...)
}

func (zl *zapLogger) Warning(args ...interface{}) {
	zl.sugar.Warn(args...)
}

func (zl *zapLogger) Warningln(args ...interface{}) {
	zl.sugar.Warn(args...)
}

func (zl *zapLogger) Warningf(format string, args ...interface{}) {
	zl.sugar.Warnf(format, args...)
}

func (zl *zapLogger) Error(args ...interface{}) {
	zl.sugar.Error(args...)
}

func (zl *zapLogger) Errorln(args ...interface{}) {
	zl.sugar.Error(args...)
}

func (zl *zapLogger) Errorf(format string, args ...interface{}) {
	zl.sugar.Errorf(format, args...)
}

func (zl *zapLogger) Fatal(args ...interface{}) {
	zl.sugar.Fatal(args...)
}

func (zl *zapLogger) Fatalln(args ...interface{}) {
	zl.sugar.Fatal(args...)
}

func (zl *zapLogger) Fatalf(format string, args ...interface{}) {
	zl.sugar.Fatalf(format, args...)
}

func (zl *zapLogger) GetLevel() string {
	return ConvertZapLevelToString(zl.cfg.Level.Level())
}
