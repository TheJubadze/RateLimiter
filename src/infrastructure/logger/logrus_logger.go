package logruslogger

import (
	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	logger *logrus.Logger
}

func NewLogrusLogger(logLevel string) *LogrusLogger {
	logger := logrus.New()
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.Fatalf("Error parsing logger level: %s", err)
	}
	logger.Printf("Logger level set to: %s", level)
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
		DisableColors:   false,
		DisableQuote:    true,
	})

	return &LogrusLogger{
		logger: logger,
	}
}

func (l *LogrusLogger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

func (l *LogrusLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}
