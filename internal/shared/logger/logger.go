package logger

import (
	"fmt"
	"io"
	"log"
	"time"
)

type Level string

const (
	INFO  Level = "INFO"
	WARN  Level = "WARN"
	ERROR Level = "ERROR"
)

type Logger struct {
	logger *log.Logger
}

func New(out io.Writer) *Logger {
	return &Logger{
		logger: log.New(out, "", 0),
	}
}

func (l *Logger) Log(level Level, msg string) {
	l.logger.Printf("%s %s %s", time.Now().Format(time.RFC3339), level, msg)
}

func (l *Logger) Info(msg string) {
	l.Log(INFO, msg)
}

func (l *Logger) Infof(format string, args ...any) {
	l.Log(INFO, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(msg string) {
	l.Log(WARN, msg)
}

func (l *Logger) Error(err error) {
	l.Log(ERROR, err.Error())
}

func (l *Logger) Errorf(format string, args ...any) {
	l.Log(ERROR, fmt.Sprintf(format, args...))
}
