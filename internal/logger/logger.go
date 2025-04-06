package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/remiehneppo/be-task-management/config"
	"github.com/sirupsen/logrus"
)

// Logger is a wrapper around logrus.Logger
type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new logger instance
func NewLogger(config *config.LoggerConfig) (*Logger, error) {
	// Create logger instance
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Configure formatter
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Setup multi writer for stdout and file
	var writers []io.Writer

	// Add console writer if enabled
	if config.EnableConsole {
		writers = append(writers, os.Stdout)
	}

	// Add file writer if enabled
	if config.EnableFile {
		if config.FilePath == "" {
			config.FilePath = "logs"
		}

		// Create logs directory if it doesn't exist
		if err := os.MkdirAll(config.FilePath, 0755); err != nil {
			return nil, fmt.Errorf("error creating log directory: %w", err)
		}

		if config.FileNamePattern == "" {
			config.FileNamePattern = "app-%Y%m%d.log"
		}

		// Set default rotation and max age if not provided
		if config.MaxAge == 0 {
			config.MaxAge = 7 * 24 * time.Hour // 7 days
		}
		if config.RotationTime == 0 {
			config.RotationTime = 24 * time.Hour // 1 day
		}

		// Initialize rotatelogs
		logPath := filepath.Join(config.FilePath, config.FileNamePattern)
		fileWriter, err := rotatelogs.New(
			logPath,
			rotatelogs.WithMaxAge(config.MaxAge),
			rotatelogs.WithRotationTime(config.RotationTime),
		)
		if err != nil {
			return nil, fmt.Errorf("error initializing log file: %w", err)
		}

		writers = append(writers, fileWriter)
	}

	// Set the multi writer for logger
	if len(writers) > 0 {
		logger.SetOutput(io.MultiWriter(writers...))
	}

	return &Logger{
		Logger: logger,
	}, nil
}

// GinLogger returns a gin.HandlerFunc middleware that logs requests using the custom logger
func (l *Logger) GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		// Get client IP and request method
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		// Format the query string
		if raw != "" {
			path = path + "?" + raw
		}

		// Determine log level based on status code
		entry := l.WithFields(logrus.Fields{
			"status":    statusCode,
			"latency":   latency,
			"client_ip": clientIP,
			"method":    method,
			"path":      path,
			// "user-agent": c.Request.UserAgent(),
		})

		msg := fmt.Sprintf("%s %s %d %s", method, path, statusCode, latency)

		if statusCode >= 500 {
			entry.Error(msg)
		} else if statusCode >= 400 {
			entry.Warn(msg)
		} else {
			entry.Info(msg)
		}
	}
}
