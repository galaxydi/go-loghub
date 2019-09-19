package sls

import (
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger = generLogger()


func generLogger() log.Logger {
	var logger log.Logger
	if LogFileName := os.Getenv("LogFileName"); LogFileName == "" {
		if IsJsonType := os.Getenv("IsJsonType"); IsJsonType == "true" {
			logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
		} else {
			logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
		}
	} else {
		if IsJsonType := os.Getenv("IsJsonType"); IsJsonType == "true" {
			logger = log.NewLogfmtLogger(initLogFlusher())
		} else {
			logger = log.NewJSONLogger(initLogFlusher())
		}
	}
	switch os.Getenv("AllowLogLevel") {
	case "debug":
		logger = level.NewFilter(logger, level.AllowDebug())
	case "info":
		logger = level.NewFilter(logger, level.AllowInfo())
	case "warn":
		logger = level.NewFilter(logger, level.AllowWarn())
	case "error":
		logger = level.NewFilter(logger, level.AllowError())
	default:
		logger = level.NewFilter(logger, level.AllowInfo())
	}

	logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	return logger
}



func initLogFlusher() *lumberjack.Logger {
	var LogMaxSize int
	var LogMaxBackups int
	if os.Getenv("LogMaxSize") == "" {
		LogMaxSize = 10
	}
	if os.Getenv("LogMaxBackups") == "" {
		LogMaxBackups = 10
	}
	return &lumberjack.Logger{
		Filename:   os.Getenv("LogFileName"),
		MaxSize:    LogMaxSize,
		MaxBackups: LogMaxBackups,
		Compress:   true,
	}
}