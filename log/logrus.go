package log

import (
	"io"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Default *logrus.Logger

type Logger = logrus.Logger

func init() {
	Default = logrus.New()
	output := &lumberjack.Logger{
		Filename:   "./logs/log.txt",
		MaxSize:    500, // megabytes
		MaxBackups: 4,
		MaxAge:     1,     // days
		Compress:   false, // disabled by default
		LocalTime:  true,
	}
	Default.SetOutput(io.MultiWriter(Default.Out, output))
	// Default.SetOutput(output)

	Default.SetLevel(logrus.DebugLevel)
}

func SetLevel(lvstr string) {
	lv, err := logrus.ParseLevel(lvstr)
	if err != nil {
		Default.Error(err)
	} else {
		Default.SetLevel(lv)
	}
}

// Tracef logs a message at level Trace on the standard logger.
func Tracef(format string, args ...interface{}) {
	Default.Tracef(format, args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	Default.Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(format string, args ...interface{}) {
	Default.Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	Default.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	Default.Warnf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	Default.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...interface{}) {
	Default.Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatalf(format string, args ...interface{}) {
	Default.Fatalf(format, args...)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	Default.Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	Default.Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	Default.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	Default.Warn(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	Default.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	Default.Panic(args...)
}
