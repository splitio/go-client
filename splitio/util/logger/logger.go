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
	_debug   log.Logger
	_info    log.Logger
	_warning log.Logger
	_error   log.Logger
}

// Debug ...
// Log a message with Debug level
func (l *Logger) Debug(msg ...interface{}) {
	l._debug.Println(msg)
}

// Info ...
// Log a message with Info level
func (l *Logger) Info(msg ...interface{}) {
	l._info.Println(msg)
}

// Warning ...
// Log a message with Warning level
func (l *Logger) Warning(msg ...interface{}) {
	l._warning.Println(msg)
}

// Error ...
// Log a message with Error level
func (l *Logger) Error(msg ...interface{}) {
	l._error.Println(msg)
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
		_debug:   *log.New(commonWriter, "DEBUG - ", 1),
		_info:    *log.New(commonWriter, "IFO - ", 1),
		_warning: *log.New(commonWriter, "WARNING - ", 1),
		_error:   *log.New(errorWriter, "ERROR - ", 1),
	}
}
