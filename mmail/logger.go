package mmail

import (
	"log"
	"os"
)

type Logger interface {
	Info(args ...interface{})
	Debug(args ...interface{})
	Error(args ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

type Log struct {
	info *log.Logger
	eror *log.Logger
	debg *log.Logger
}

type devNull int

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

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

func (l *Log) Info(args ...interface{}) {
	l.info.Println(args)
}

func (l *Log) Debug(args ...interface{}) {
	l.debg.Println(args)
}

func (l *Log) Error(args ...interface{}) {
	l.eror.Println(args)
}

func (l *Log) Infof(format string, v ...interface{}) {
	l.info.Printf(format, v)
}

func (l *Log) Debugf(format string, v ...interface{}) {
	l.debg.Printf(format, v)
}

func (l *Log) Errorf(format string, v ...interface{}) {
	l.eror.Printf(format, v)
}
