package db

import (
	"github.com/yyklll/skeleton/pkg/log"

	"xorm.io/core"
)

// XORMLogBridge a logger bridge from Logger to xorm
type XORMLogBridge struct {
	showSQL bool
}

// NewXORMLogger inits a log bridge for xorm
func NewXORMLogger(showSQL bool) core.ILogger {
	return &XORMLogBridge{
		showSQL: showSQL,
	}
}

// Debug show debug log
func (l *XORMLogBridge) Debug(v ...interface{}) {
	log.Debug(v...)
}

// Debugf show debug log
func (l *XORMLogBridge) Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

// Error show error log
func (l *XORMLogBridge) Error(v ...interface{}) {
	log.Error(v...)
}

// Errorf show error log
func (l *XORMLogBridge) Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

// Info show information level log
func (l *XORMLogBridge) Info(v ...interface{}) {
	log.Info(v...)
}

// Infof show information level log
func (l *XORMLogBridge) Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// Warn show warning log
func (l *XORMLogBridge) Warn(v ...interface{}) {
	log.Warning(v...)
}

// Warnf show warnning log
func (l *XORMLogBridge) Warnf(format string, v ...interface{}) {
	log.Warningf(format, v...)
}

// Level get logger level
func (l *XORMLogBridge) Level() core.LogLevel {
	switch log.GetLevel() {
	case "debug":
		return core.LOG_DEBUG
	case "info":
		return core.LOG_INFO
	case "warn":
		return core.LOG_WARNING
	default:
		return core.LOG_ERR
	}
	return core.LOG_OFF
}

// SetLevel set the logger level
func (l *XORMLogBridge) SetLevel(lvl core.LogLevel) {
}

// ShowSQL set if record SQL
func (l *XORMLogBridge) ShowSQL(show ...bool) {
	if len(show) > 0 {
		l.showSQL = show[0]
	} else {
		l.showSQL = true
	}
}

// IsShowSQL if record SQL
func (l *XORMLogBridge) IsShowSQL() bool {
	return l.showSQL
}
