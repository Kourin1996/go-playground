package log

import "github.com/labstack/gommon/log"

type ILoggerManager interface {
	GetLoggers(level Level) []*log.Logger
}

type Level uint8

const (
	PRINT Level = iota
	DEBUG
	INFO
	WARN
	ERROR
	OFF
	PANIC
	FATAL
)
