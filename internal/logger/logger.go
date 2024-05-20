package logger

import (
	"github.com/sirupsen/logrus"
)

func New() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	// logger.SetLevel(logrus.InfoLevel)
	return logger
}

func NewTextFormat() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	// logger.SetLevel(logrus.InfoLevel)
	return logger
}
