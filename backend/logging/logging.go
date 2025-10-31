package logging

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
)

const (
	colorReset  = "\033[0m"
	colorGray   = "\033[90m"
	colorWhite  = "\033[97m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
)

type Logger struct {
	logger             *log.Logger
	JSONLoggerInstance *slog.Logger
	logLevel           LogLevel
	LoggingEnabled     bool
	Ansi               bool
	JSON               bool
}

var (
	instance *Logger
	once     sync.Once
)

func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			logLevel:           INFO,
			LoggingEnabled:     true,
			logger:             log.New(os.Stdout, "", log.LstdFlags),
			JSON:               false,
			JSONLoggerInstance: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
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

func (l *Logger) SetJSON(json bool) {
	l.JSON = json
}

func (l *Logger) SetAnsi(ansi bool) {
	l.Ansi = ansi
}

func (l *Logger) log(level, color, message string, msgLevel LogLevel) {
	if !l.JSON {
		if l.Ansi {
			l.logger.Printf("%s%s%s%s", color, level, message, colorReset)
		} else {
			l.logger.Printf("%s%s", level, message)
		}
	} else {
		switch msgLevel {
		case DEBUG:
			l.JSONLoggerInstance.Debug(message)
		case INFO:
			l.JSONLoggerInstance.Info(message)
		case WARNING:
			l.JSONLoggerInstance.Warn(message)
		case ERROR:
			l.JSONLoggerInstance.Error(message)
		default:
			l.JSONLoggerInstance.Info(message)
		}
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
		l.log("[DEBUG] ", colorGray, message, DEBUG)
	} else {
		l.log("[DEBUG] ", colorGray, format, DEBUG)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if !l.verifyLog(INFO) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[INFO] ", colorWhite, message, INFO)
	} else {
		l.log("[INFO] ", colorWhite, format, INFO)
	}
}

func (l *Logger) Warning(format string, args ...interface{}) {
	if !l.verifyLog(WARNING) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[WARN] ", colorYellow, message, WARNING)
	} else {
		l.log("[WARN] ", colorYellow, format, WARNING)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if !l.verifyLog(ERROR) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[ERROR] ", colorRed, message, ERROR)
	} else {
		l.log("[ERROR] ", colorRed, format, ERROR)
	}
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	if !l.verifyLog(FATAL) {
		return
	}
	if len(args) > 0 {
		message := fmt.Sprintf(format, args...)
		l.log("[FATAL] ", colorRed, message, FATAL)
	} else {
		l.log("[FATAL] ", colorRed, format, FATAL)
	}
	os.Exit(1)
}
