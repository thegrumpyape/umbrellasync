package logging

import (
	"io"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type FileHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
	Formatter logrus.Formatter
}

func (hook *FileHook) Levels() []logrus.Level {
	return hook.LogLevels
}

func (hook *FileHook) Fire(entry *logrus.Entry) error {
	data, err := hook.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write(data)
	return err
}

func NewFileHook(writer io.Writer, formatter logrus.Formatter) *FileHook {
	return &FileHook{
		Writer:    writer,
		Formatter: formatter,
		LogLevels: logrus.AllLevels,
	}
}
