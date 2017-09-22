// Package logging ...
// Handles logging within the SDK
package logging

import (
	"io"
	"log"
	"os"
)

// LoggerOptions ...
// Struct that must be passed to the NewLogger constructor to setup a logger
// CommonWriter and ErrorWriter can be <nil>. In that case they'll default to os.Stdout
type LoggerOptions struct {
	CommonWriter io.Writer
	ErrorWriter  io.Writer
}

// Logger struct. Encapsulates four different loggers, each for a different "level",
// and provides Error, Debug, Warning and Info functions, that will forward a message
// to the appropriate logger.
type Logger struct {
	debugLogger   log.Logger
	infoLogger    log.Logger
	warningLogger log.Logger
	errorLogger   log.Logger
}

// Debug ...
// Log a message with Debug level
func (l *Logger) Debug(msg ...interface{}) {
	l.debugLogger.Println(msg)
}

// Info ...
// Log a message with Info level
func (l *Logger) Info(msg ...interface{}) {
	l.infoLogger.Println(msg)
}

// Warning ...
// Log a message with Warning level
func (l *Logger) Warning(msg ...interface{}) {
	l.warningLogger.Println(msg)
}

// Error ...
// Log a message with Error level
func (l *Logger) Error(msg ...interface{}) {
	l.errorLogger.Println(msg)
}

// NewLogger ...
// Instantiates a new Logger instance.
// Requires a pointer to a LoggerOptions struct to be passed.
func NewLogger(options *LoggerOptions) *Logger {
	errorWriter := options.ErrorWriter
	commonWriter := options.CommonWriter

	if options.ErrorWriter == nil {
		errorWriter = os.Stdout
	}

	if options.CommonWriter == nil {
		commonWriter = os.Stdout
	}

	return &Logger{
		debugLogger:   *log.New(commonWriter, "DEBUG - ", 1),
		infoLogger:    *log.New(commonWriter, "INFO - ", 1),
		warningLogger: *log.New(commonWriter, "WARNING - ", 1),
		errorLogger:   *log.New(errorWriter, "ERROR - ", 1),
	}
}
