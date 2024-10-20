// internal/logger/logger.go

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DebugLevel is used for detailed system operations
	DebugLevel LogLevel = iota
	// InfoLevel is used for general operational entries
	InfoLevel
	// WarningLevel is used for non-critical issues
	WarningLevel
	// ErrorLevel is used for errors that need attention
	ErrorLevel
)

var (
	instance *Logger
	once     sync.Once
)

// Logger handles all logging operations
type Logger struct {
	level  LogLevel
	logger *log.Logger
	file   *os.File
	mu     sync.Mutex
}

// GetLogger returns the singleton instance of Logger
func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			level:  InfoLevel,
			logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
		}
		instance.setupLogFile()
	})
	return instance
}

func (l *Logger) setupLogFile() {
	logDir := config.GetConfig().LogDirectory
	if logDir == "" {
		logDir = "logs"
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		l.Error(i18n.GetMessage("ErrCreateLogDir"), err)
		return
	}

	logFile := filepath.Join(logDir, fmt.Sprintf("ontology_%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		l.Error(i18n.GetMessage("ErrOpenLogFile"), err)
		return
	}

	l.file = file
	l.logger.SetOutput(io.MultiWriter(os.Stdout, file))
}

// SetLevel sets the current log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, file, line, _ := runtime.Caller(2)
	levelStr := [...]string{"DEBUG", "INFO", "WARNING", "ERROR"}[level]
	logMessage := fmt.Sprintf("[%s] %s:%d - %s", levelStr, filepath.Base(file), line, fmt.Sprintf(message, args...))
	l.logger.Println(logMessage)
}

// Debug logs a message at DebugLevel
func (l *Logger) Debug(format string, args ...interface{}) {
    if l.GetLevel() <= DebugLevel {
        l.log(DebugLevel, format, args...)
    }
}

// Info logs a message at InfoLevel
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(InfoLevel, message, args...)
}

// Warning logs a message at WarningLevel
func (l *Logger) Warning(message string, args ...interface{}) {
	l.log(WarningLevel, message, args...)
}

// Error logs a message at ErrorLevel
func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ErrorLevel, message, args...)
}

// Close closes the log file
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

// UpdateProgress updates the progress on the console
func (l *Logger) UpdateProgress(current, total int) {
	fmt.Printf("\rProgress: %d/%d", current, total)
}

// RotateLogs archives old log files
func (l *Logger) RotateLogs() error {
	// Implementation of log rotation
	// This is a placeholder and should be implemented based on specific requirements
	return nil
}

// ParseLevel converts a string level to LogLevel
func ParseLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warning":
		return WarningLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// Ajoutez cette méthode
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}
