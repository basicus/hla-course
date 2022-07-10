package log

import (
	"github.com/sirupsen/logrus"
)

func New(cfg Config) *logrus.Logger {
	logger := logrus.StandardLogger()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: cfg.DisableTimestamp,
	})

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		logger.WithError(err).WithField("level", cfg.Level).Warn("Cannot parse a logging level")
	} else {
		logger.SetLevel(level)
	}

	return logger
}
