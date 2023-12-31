package logger

import (
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
)

func New() *logrus.Logger {
	logger := logrus.New()
	fileHook := lfshook.NewHook(lfshook.PathMap{
		logrus.InfoLevel: "logger/logfile.log",
	}, &logrus.JSONFormatter{})
	logger.AddHook(fileHook)
	logger.SetOutput(os.Stdout)
	return logger
}
