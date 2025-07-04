package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

// L 是全局的 logrus logger 实例
var L *logrus.Logger

func init() {
	// 创建新的 logger 实例
	L = logrus.New()

	// 设置输出到标准输出
	L.SetOutput(os.Stdout)

	// 设置日志格式为文本格式，带颜色
	L.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 默认设置为 INFO 级别，可以通过环境变量 LOG_LEVEL 来调整
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		L.SetLevel(logrus.DebugLevel)
	case "WARN":
		L.SetLevel(logrus.WarnLevel)
	case "ERROR":
		L.SetLevel(logrus.ErrorLevel)
	case "FATAL":
		L.SetLevel(logrus.FatalLevel)
	default:
		L.SetLevel(logrus.InfoLevel)
	}
}

// SetDebugLevel 设置日志级别为 DEBUG
func SetDebugLevel() {
	L.SetLevel(logrus.DebugLevel)
}

// SetInfoLevel 设置日志级别为 INFO
func SetInfoLevel() {
	L.SetLevel(logrus.InfoLevel)
}
