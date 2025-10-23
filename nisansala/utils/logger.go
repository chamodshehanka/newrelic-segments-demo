package utils

import (
	"go.uber.org/zap"
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

var currentLogLevel = INFO

type CustomLogger struct {
	*zap.SugaredLogger
}

var Logger *CustomLogger

func init() {
	rawLogger, _ := zap.NewProduction()
	defer rawLogger.Sync()
	Logger = &CustomLogger{rawLogger.Sugar()}
}

func SetLogLevel(level int) {
	currentLogLevel = level
	switch level {
	case DEBUG:
		Logger.Infof("Logger initialized successfully with log level: DEBUG")
	case INFO:
		Logger.Infof("Logger initialized successfully with log level: INFO")
	case ERROR:
		Logger.Infof("Logger initialized successfully with log level: ERROR")
	default:
		panic("unhandled default case")
	}
}

func (logger *CustomLogger) Debug(requestID, msg string, args ...interface{}) {
	if currentLogLevel <= DEBUG {
		logger.With("requestID", requestID).Debugf(msg, args...)
	}
}

func (logger *CustomLogger) Info(requestID, msg string, args ...interface{}) {
	if currentLogLevel <= INFO {
		logger.With("requestID", requestID).Infof(msg, args...)
	}
}

func (logger *CustomLogger) Error(requestID, msg string, args ...interface{}) {
	if currentLogLevel <= ERROR {
		logger.With("requestID", requestID).Errorf(msg, args...)
	}
}
