// logger/logger.go
package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

type Logger struct {
	mu       sync.Mutex
	logLevel LogLevel
	logger   *log.Logger
}

var (
	instance *Logger
	once     sync.Once
)

func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			logLevel: INFO,
			logger:   log.New(os.Stdout, "", log.LstdFlags),
		}
	})
	return instance
}

func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logLevel = level
}

func (l *Logger) log(levelStr string, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("%s%s", levelStr, message)
}

func (l *Logger) verifyLogLevel(level LogLevel) bool {
	return level >= l.logLevel
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.verifyLogLevel(DEBUG) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[DEBUG] ", message)
	} else {
		l.log("[DEBUG] ", format)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if !l.verifyLogLevel(INFO) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[INFO] ", message)
	} else {
		l.log("[INFO] ", format)
	}
}

func (l *Logger) Warning(format string, args ...interface{}) {
	if !l.verifyLogLevel(WARNING) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[WARN] ", message)
	} else {
		l.log("[WARN] ", format)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if !l.verifyLogLevel(ERROR) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[ERROR] ", message)
	} else {
		l.log("[ERROR] ", format)
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

func ToLogLevel(logLevel int) LogLevel {
	switch logLevel {
	case 0:
		return DEBUG
	case 1:
		return INFO
	case 2:
		return WARNING
	case 3:
		return ERROR
	default:
		return INFO
	}
}

func ToInteger(logLevel LogLevel) int {
	switch logLevel {
	case DEBUG:
		return 0
	case INFO:
		return 1
	case WARNING:
		return 2
	case ERROR:
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
