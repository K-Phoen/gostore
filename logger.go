package gostore

import (
	"github.com/sirupsen/logrus"
)

type prefixedFormatter struct {
	prefix    string
	formatter logrus.Formatter
}

func (f prefixedFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := []byte(f.prefix)

	formatted, err := f.formatter.Format(entry)
	if err != nil {
		return nil, err
	}

	return append(b, formatted...), nil
}

func newPrefixedLogger(logger *logrus.Logger, prefix string) *logrus.Logger {
	newLogger := logrus.New()
	newLogger.Out = logger.Out
	newLogger.SetLevel(logger.Level)
	newLogger.SetFormatter(prefixedFormatter{
		prefix:    prefix,
		formatter: logger.Formatter,
	})

	return newLogger
}
