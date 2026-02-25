package logger

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	gormlogger "gorm.io/gorm/logger"
)

var Log *logrus.Logger

func init() {
	godotenv.Load()
	Log = logrus.New()

	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
		DisableLevelTruncation: true,
		PadLevelText:           true,
	})

	Log.SetOutput(os.Stdout)

	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch logLevel {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		Log.SetLevel(logrus.FatalLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

func WithField(key string, value any) *logrus.Entry {
	return Log.WithField(key, value)
}

func NewGormLogger() gormlogger.Interface {
	var gormLogLevel gormlogger.LogLevel

	switch Log.Level {
	case logrus.DebugLevel:
		gormLogLevel = gormlogger.Info
	case logrus.WarnLevel:
		gormLogLevel = gormlogger.Warn
	case logrus.ErrorLevel, logrus.FatalLevel:
		gormLogLevel = gormlogger.Error
	default:
		gormLogLevel = gormlogger.Silent
	}

	return gormlogger.New(
		Log,
		gormlogger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      gormLogLevel,
			Colorful:      false,
		},
	)
}
