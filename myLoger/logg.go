package logg

import (
	"fmt"
	"log"
	"os"
)

const (
	TRACE = iota
	DEBUG
	WARN
	INFO
	ERROR
	FATAL
)

const (
	trace = "[TRACE]"
	debug = "[DEBUG]"
	warn  = "[WARN]"
	info  = "[INFO]"
	error = "[ERROR]"
	fatal = "[FATAL]"
)

type ModuleLogger struct {
	l          *log.Logger
	moduleName string
	enabled    bool
	level      int
}

func (l *ModuleLogger) Debug(format string, i ...interface{}) {
	l.log(DEBUG, format, i...)
}

func (l *ModuleLogger) Error(format string, i ...interface{}) {
	l.log(ERROR, format, i...)
}

func (l *ModuleLogger) log(level int, format string, i ...interface{}) {
	if l.enabled && l.level <= level {
		err := l.l.Output(3, "{"+l.moduleName+"} "+debug+fmt.Sprintf(format, i...))
		if err != nil {
			panic("Logging failed with " + err.Error())
		}
	}
}

type LogManager struct {
	l       *log.Logger
	modules map[string]*ModuleLogger
}

func (lm *LogManager) RegisterModule(name *string) *ModuleLogger {
	lm.modules[*name] = &ModuleLogger{l: lm.l, level: FATAL, enabled: false}
	return lm.modules[*name]
}

var DefaultModuleLogger = &ModuleLogger{enabled: true, level: DEBUG, moduleName: "DefaultModule",
	l: log.New(os.Stdout, "", log.LstdFlags|log.Ltime|log.Lshortfile)}
