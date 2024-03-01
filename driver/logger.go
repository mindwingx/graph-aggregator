package driver

import (
	"fmt"
	aggregator "github.com/mindwingx/graph-aggregator"
	"github.com/mindwingx/graph-aggregator/driver/abstractions"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type (
	Zap struct {
		logger *zap.Logger
		conf   loggerConfig
	}

	loggerConfig struct {
		LogFileName string `mapstructure:"LOG_FILE_NAME"`
	}
)

func InitLogger(registry abstractions.RegAbstraction) abstractions.LoggerAbstraction {
	l := new(Zap)
	registry.Parse(&l.conf)

	// instantiate
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	// log file configs
	fileEncoder := zapcore.NewJSONEncoder(config)
	// consoleEncoder := zapcore.NewConsoleEncoder(config) // todo: uncomment to print in STDOUT
	defaultLogLevel := zapcore.InfoLevel
	// log file instantiate
	logFilePath := fmt.Sprintf("%s/logs/%s", aggregator.Root(), l.conf.LogFileName)
	logFile, _ := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	// zap core
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		// zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel), // todo: uncomment to print in STDOUT
	)

	l.logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return l
}

func (z *Zap) Logger() *zap.Logger {
	return z.logger
}
