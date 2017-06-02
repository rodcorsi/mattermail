package mmail

import (
	"log"
	"os"
)

// Logger default interface logger with info/debug/error
type Logger interface {
	Info(args ...interface{})
	Debug(args ...interface{})
	Error(args ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

// Log implements Logger interface
type Log struct {
	info *log.Logger
	eror *log.Logger
	debg *log.Logger
}

type devNull int

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

// NewLog creates a new Logger
func NewLog(prefix string, debug bool) *Log {
	logger := &Log{
		info: log.New(os.Stdout, "INFO "+prefix+"\t", log.Ltime),
		eror: log.New(os.Stderr, "EROR "+prefix+"\t", log.Ltime),
	}

	if debug {
		logger.debg = log.New(os.Stdout, "DEBG "+prefix+"\t", log.Ltime)
	} else {
		logger.debg = log.New(devNull(0), "", 0)
	}

	return logger
}

// Info calls Println with tag INFO
func (l *Log) Info(args ...interface{}) {
	l.info.Println(args)
}

// Debug calls Println with tag DEBG
func (l *Log) Debug(args ...interface{}) {
	l.debg.Println(args)
}

// Error calls Println with tag EROR
func (l *Log) Error(args ...interface{}) {
	l.eror.Println(args)
}

// Infof calls Printf with tag INFO
func (l *Log) Infof(format string, v ...interface{}) {
	l.info.Printf(format, v)
}

// Debugf calls Printf with tag DEBG
func (l *Log) Debugf(format string, v ...interface{}) {
	l.debg.Printf(format, v)
}

// Errorf calls Printf with tag EROR
func (l *Log) Errorf(format string, v ...interface{}) {
	l.eror.Printf(format, v)
}
