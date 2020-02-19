package log

import (
	"sync"
	"time"

	"github.com/yyklll/skeleton/pkg/util"
)

var DefaultLogLevel = "info"

// global logger
var _gl Logger
var _gl_once sync.Once

func Debug(args ...interface{})                   { _gl.Info(args...) }
func Debugln(args ...interface{})                 { _gl.Info(args...) }
func Debugf(format string, args ...interface{})   { _gl.Infof(format, args...) }
func Info(args ...interface{})                    { _gl.Info(args...) }
func Infoln(args ...interface{})                  { _gl.Info(args...) }
func Infof(format string, args ...interface{})    { _gl.Infof(format, args...) }
func Warning(args ...interface{})                 { _gl.Warning(args...) }
func Warningln(args ...interface{})               { _gl.Warning(args...) }
func Warningf(format string, args ...interface{}) { _gl.Warningf(format, args...) }
func Error(args ...interface{})                   { _gl.Error(args...) }
func Errorln(args ...interface{})                 { _gl.Error(args...) }
func Errorf(format string, args ...interface{})   { _gl.Errorf(format, args...) }
func Fatal(args ...interface{})                   { _gl.Fatal(args...) }
func Fatalln(args ...interface{})                 { _gl.Fatal(args...) }
func Fatalf(format string, args ...interface{})   { _gl.Fatalf(format, args...) }
func GetLevel() string                            { return _gl.GetLevel() }

// Logger defines logging interface.
type Logger interface {
	Debug(args ...interface{})
	Debugln(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})
	Warning(args ...interface{})
	Warningln(args ...interface{})
	Warningf(format string, args ...interface{})
	Error(args ...interface{})
	Errorln(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalln(args ...interface{})
	Fatalf(format string, args ...interface{})
	GetLevel() string
	// SetLogLevelAtRuntime(level string)
}

// assert that "defaultLogger" satisfy "Logger" interface
var _ Logger = &zapLogger{}

func InitGlobalLogger(lvl string) {
	if l, err := NewZapLogger(InitZapConfig(lvl)); err == nil {
		ReplaceGlobalLogger(l)
		return
	}
	panic("global logger initialized error, exit program")
}

func GetGlobalLogger() *Logger {
	return &_gl
}

// It should only be called once
// So the global default logger couldn't be replaced once setup
func ReplaceGlobalLogger(logger Logger) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		_gl_once.Do(func() {
			_gl = logger
		})
		wg.Done()
	}()

	if util.WaitTimeout(&wg, 10*time.Second) {
		panic("global logger took too long to initialize")
	}
}
