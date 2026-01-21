package logger

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	log  *logrus.Logger
	once sync.Once
)

// Init initializes the logger only once
func Init() {
	once.Do(func() {
		l := logrus.New()

		// Output
		l.SetOutput(os.Stdout)

		// Formatter (JSON recommended for prod)
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})

		// Log level (can be env-driven)
		level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
		if err != nil {
			level = logrus.InfoLevel
		}
		l.SetLevel(level)

		// Enable caller info if needed
		// l.SetReportCaller(true)

		log = l
	})
}

// GetLogger returns the singleton logger
func GetLogger() *logrus.Logger {
	if log == nil {
		Init()
	}
	return log
}

func Info(msg string) {
	GetLogger().Info(msg)
}

func Error(err error, msg string) {
	GetLogger().WithError(err).Error(msg)
}

func Fatal(err error, msg string) {
	GetLogger().WithError(err).Fatal(msg)
}
