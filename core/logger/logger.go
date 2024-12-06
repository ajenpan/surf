package logger

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ajenpan/surf/core/utils/logrotate"
)

var Default = New(Options{
	Level:      slog.LevelDebug,
	RotateTime: time.Hour * 24,
	LogPrefix:  "server",
})

type Options struct {
	Level      slog.Level
	RotateTime time.Duration
	LogPrefix  string
}

type Logger struct {
	level *slog.LevelVar
	impl  *slog.Logger
}

func New(opts Options) *Logger {
	logpath := fmt.Sprintf("./log/%s.20060102150405.log", opts.LogPrefix)
	var logLevel = &slog.LevelVar{}
	rl, err := logrotate.NewRoteteLog(logpath, logrotate.WithRotateTime(opts.RotateTime))
	if err != nil {
		panic(err)
	}
	var loghandler = slog.NewTextHandler(rl, &slog.HandlerOptions{Level: opts.Level})
	log := slog.New(loghandler)
	return &Logger{
		level: logLevel,
		impl:  log,
	}
}

func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}

func (l *Logger) Slog() *slog.Logger {
	return l.impl
}
