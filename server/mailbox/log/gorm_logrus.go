package log

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type GormLogrus struct {
	Log                   *log.Logger
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
	LogLevel              gormlogger.LogLevel
	// SkipCallerLookup      bool
}

func NewGormLogrus() *GormLogrus {
	return &GormLogrus{
		Log:                   Default.WithField("component", "gorm").Logger,
		SkipErrRecordNotFound: false,
		LogLevel:              gormlogger.Info,
		SlowThreshold:         100 * time.Millisecond,
	}
}

func (l *GormLogrus) LogMode(lv gormlogger.LogLevel) gormlogger.Interface {
	ret := *l
	ret.LogLevel = lv
	//this is not cpp's code
	return &ret
}

func (l *GormLogrus) Info(ctx context.Context, s string, args ...interface{}) {
	l.Log.WithContext(ctx).Infof(s, args...)
}

func (l *GormLogrus) Warn(ctx context.Context, s string, args ...interface{}) {
	l.Log.WithContext(ctx).Warnf(s, args...)
}

func (l *GormLogrus) Error(ctx context.Context, s string, args ...interface{}) {
	l.Log.WithContext(ctx).Errorf(s, args...)
}

func (l *GormLogrus) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	fields := log.Fields{}
	if l.SourceField != "" {
		fields[l.SourceField] = utils.FileWithLineNum()
	}

	switch {
	case err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound):
		sql, _ := fc()
		fields[log.ErrorKey] = err
		l.Log.WithContext(ctx).WithFields(fields).Errorf("[%s] %s ", elapsed, sql)
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold:
		sql, _ := fc()
		l.Log.WithContext(ctx).WithFields(fields).Warnf("[slow sql] [%s] %s", elapsed, sql)
	case l.LogLevel == gormlogger.Info:
		sql, _ := fc()
		l.Log.WithContext(ctx).WithFields(fields).Infof("[%s] %s", elapsed, sql)
	}
}
