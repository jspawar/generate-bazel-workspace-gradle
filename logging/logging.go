package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	// enable colorized log levels
	logEncConfig := zap.NewDevelopmentEncoderConfig()
	logEncConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logConfig := zap.NewDevelopmentConfig()
	logConfig.EncoderConfig = logEncConfig

	// replace global zap logger
	globalLogger, _ := logConfig.Build()
	zap.ReplaceGlobals(globalLogger)
}
