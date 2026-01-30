package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

func NewLogger(debugEnabled, infoEnabled bool, errorOutput, infoOutput, debugOutput string) *Logger {
	return &Logger{
		errorLogger:  log.New(openOutput(errorOutput, os.Stderr), "", 0),
		infoLogger:   log.New(openOutput(infoOutput, os.Stdout), "", 0),
		debugLogger:  log.New(openOutput(debugOutput, os.Stdout), "", 0),
		debugEnabled: debugEnabled,
		infoEnabled:  infoEnabled,
	}
}

// openOutput opens the specified output destination.
// Supports "stdout", "stderr", or a file path.
func openOutput(path string, defaultOutput *os.File) *os.File {
	switch path {
	case "", "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	default:
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open log file %s: %v, using default output", path, err)
			return defaultOutput
		}
		return file
	}
}

type Logger struct {
	errorLogger  *log.Logger
	infoLogger   *log.Logger
	debugLogger  *log.Logger
	debugEnabled bool
	infoEnabled  bool
}

func (l *Logger) Debugf(format string, v ...any) {
	if l == nil || !l.debugEnabled {
		return
	}
	timestamp := time.Now().Format("2006-01-02T15:04:05.000000-07:00")
	l.debugLogger.Printf("%s [DEBUG] %s",
		timestamp, fmt.Sprintf(format, v...))
}

func (l *Logger) Infof(format string, v ...any) {
	if l == nil || !l.infoEnabled {
		return
	}
	timestamp := time.Now().Format("2006-01-02T15:04:05.000000-07:00")
	l.infoLogger.Printf("%s [INFO] %s",
		timestamp, fmt.Sprintf(format, v...))
}

func (l *Logger) Warnf(format string, v ...any) {
	if l == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02T15:04:05.000000-07:00")
	l.errorLogger.Printf("%s [WARN] %s",
		timestamp, fmt.Sprintf(format, v...))
}

func (l *Logger) Errorf(format string, v ...any) {
	if l == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02T15:04:05.000000-07:00")
	l.errorLogger.Printf("%s [ERROR] %s",
		timestamp, fmt.Sprintf(format, v...))
}