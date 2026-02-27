package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	gormlogger "gorm.io/gorm/logger"
)

var logLevel string

func init() {
	godotenv.Load()

	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(os.Stdout)

	logLevel = strings.ToLower(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "info"
	}
}

func shouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}

	current, ok := levels[logLevel]
	if !ok {
		current = 1 // default to info
	}

	target, ok := levels[level]
	if !ok {
		target = 1
	}

	return target >= current
}

func logMessage(level, format string, args ...any) {
	if !shouldLog(level) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	log.Printf("[%s] %s", strings.ToUpper(level), msg)
}

func Debug(msg string) {
	logMessage("debug", "%s", msg)
}

func Info(msg string) {
	logMessage("info", "%s", msg)
}

func Warn(msg string) {
	logMessage("warn", "%s", msg)
}

func Error(msg string) {
	logMessage("error", "%s", msg)
}

type gormLogWriter struct{}

func (w gormLogWriter) Printf(format string, args ...any) {
	logMessage("info", format, args...)
}

func NewGormLogger() gormlogger.Interface {
	var gormLogLevel gormlogger.LogLevel

	switch logLevel {
	case "debug":
		gormLogLevel = gormlogger.Info
	case "warn", "warning":
		gormLogLevel = gormlogger.Warn
	case "error":
		gormLogLevel = gormlogger.Error
	default:
		gormLogLevel = gormlogger.Silent
	}

	return gormlogger.New(
		gormLogWriter{},
		gormlogger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      gormLogLevel,
			Colorful:      false,
		},
	)
}
