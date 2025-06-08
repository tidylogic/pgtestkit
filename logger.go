package pgtestkit

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// privateLogger is the singleton logger instance for the package
	privateLogger *zap.Logger
	onceLogger    sync.Once

	// enableLogging controls whether logging is enabled
	enableLogging = false
)

// SetLogging enables or disables logging
func SetLogging(enabled bool) {
	enableLogging = enabled
}

// IsLoggingEnabled returns whether logging is currently enabled
func IsLoggingEnabled() bool {
	return enableLogging
}

// getLogger returns a thread-safe singleton logger instance
func getLogger() *zap.Logger {
	onceLogger.Do(func() {
		// Configure the logger with production settings
		config := zap.NewProductionConfig()

		// Customize the encoder config for better readability
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		config.EncoderConfig = encoderConfig

		// Set the output to stderr by default
		config.OutputPaths = []string{"stderr"}

		// In development, use a more human-friendly console encoder
		if os.Getenv("ENV") == "development" {
			config = zap.NewDevelopmentConfig()
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}

		// Build the logger
		var err error
		privateLogger, err = config.Build()
		if err != nil {
			// Fallback to a basic logger if there's an error
			privateLogger = zap.NewExample()
		}

		// Add caller information to log entries
		privateLogger = privateLogger.WithOptions(zap.AddCaller())

		// Set the global logger
		zap.ReplaceGlobals(privateLogger)
	})

	return privateLogger
}

// logError logs an error message with additional context
func logError(msg string, err error, fields ...zap.Field) {
	if !enableLogging {
		return
	}
	allFields := append([]zap.Field{zap.Error(err)}, fields...)
	getLogger().Error(msg, allFields...)
}

// logInfo logs an info message with additional context
func logInfo(msg string, fields ...zap.Field) {
	if !enableLogging {
		return
	}
	getLogger().Info(msg, fields...)
}

// logDebug logs a debug message with additional context
func logDebug(msg string, fields ...zap.Field) {
	if !enableLogging {
		return
	}
	getLogger().Debug(msg, fields...)
}

// logWarn logs a warning message with additional context
func logWarn(msg string, fields ...zap.Field) {
	if !enableLogging {
		return
	}
	getLogger().Warn(msg, fields...)
}
