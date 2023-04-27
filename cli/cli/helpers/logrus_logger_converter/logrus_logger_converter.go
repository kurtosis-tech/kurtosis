package logrus_logger_converter

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
)

type logrusLoggerImpl struct {
	logrus *logrus.Logger
}

func (logger *logrusLoggerImpl) Logf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func (logger *logrusLoggerImpl) Errorf(format string, args ...interface{}) {
	logger.logrus.Errorf(format, args...)
}

func ConvertLogrusLoggerToAnalyticsLogger(logger *logrus.Logger) analytics.Logger {
	return &logrusLoggerImpl{
		logrus: logger,
	}
}
