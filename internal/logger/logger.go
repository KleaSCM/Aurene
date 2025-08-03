/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: logger.go
Description: Thread-safe logging system for the Aurene scheduler. Provides different log levels
(DEBUG, INFO, WARN, ERROR, FATAL) with timestamped output for debugging and monitoring.
*/

package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

/**
 * LogLevel represents the logging level hierarchy
 *
 * Provides structured logging levels for different message types.
 * Enables runtime log filtering and severity-based output control.
 */
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

/**
 * Logger provides thread-safe logging functionality
 *
 * Implements concurrent logging with level filtering and
 * timestamp formatting for scheduler debugging and monitoring.
 */
type Logger struct {
	mu    sync.Mutex
	level LogLevel
}

/**
 * New creates a new logger instance with default INFO level
 *
 * Returns a thread-safe logger ready for scheduler operations.
 */
func New() *Logger {
	return &Logger{
		level: INFO,
	}
}

/**
 * SetLevel configures the logging level for runtime filtering
 *
 * Allows dynamic adjustment of log verbosity during scheduler operation.
 */
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

/**
 * log formats and outputs messages with level filtering and timestamping
 *
 * Implements the core logging logic with thread-safe output formatting.
 * Only outputs messages that meet or exceed the configured log level.
 */
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := l.levelToString(level)
	message := fmt.Sprintf(format, args...)

	log.Printf("[%s] %s: %s", timestamp, levelStr, message)
}

func (l *Logger) levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}
