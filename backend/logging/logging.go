package logging

import (
	"fmt"
	"log"
	"os"
	"sync"
)

const (
	ColorReset  = "\033[0m"
	ColorGray   = "\033[90m"
	ColorWhite  = "\033[97m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

type Logger struct {
	logLevel       LogLevel
	LoggingEnabled bool
	Ansi           bool
	logger         *log.Logger
}

var (
	instance *Logger
	once     sync.Once
)

func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			logLevel:       INFO,
			LoggingEnabled: true,
			logger:         log.New(os.Stdout, "", log.LstdFlags),
		}
	})
	return instance
}

func (l *Logger) ToggleLogging(logging bool) {
	if l.LoggingEnabled && !logging {
		l.Info("Logging is being disabled. Go to admin panel -> Settings to turn it back on.")
	}
	l.LoggingEnabled = logging
}

func (l *Logger) SetLevel(level LogLevel) {
	if l.LoggingEnabled {
		l.logLevel = level
	}
}

func (l *Logger) log(level string, color string, message string) {
	if l.Ansi {
		l.logger.Printf("%s%s%s%s", color, level, message, ColorReset)
	} else {
		l.logger.Printf("%s%s", level, message)
	}
}

func (l *Logger) verifyLog(level LogLevel) bool {
	return level >= l.logLevel && l.LoggingEnabled
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.verifyLog(DEBUG) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[DEBUG] ", ColorGray, message)
	} else {
		l.log("[DEBUG] ", ColorGray, format)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if !l.verifyLog(INFO) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[INFO] ", ColorWhite, message)
	} else {
		l.log("[INFO] ", ColorWhite, format)
	}
}

func (l *Logger) Warning(format string, args ...interface{}) {
	if !l.verifyLog(WARNING) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[WARN] ", ColorYellow, message)
	} else {
		l.log("[WARN] ", ColorYellow, format)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if !l.verifyLog(ERROR) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[ERROR] ", ColorRed, message)
	} else {
		l.log("[ERROR] ", ColorRed, format)
	}
}

func FromString(logLevel string) LogLevel {
	switch logLevel {
	case "DEBUG":
		return 0
	case "INFO":
		return 1
	case "WARNING":
		return 2
	case "ERROR":
		return 3
	default:
		return 1
	}
}

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
