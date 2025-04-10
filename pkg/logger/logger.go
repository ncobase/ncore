package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ncobase/ncore/pkg/config"
	"github.com/ncobase/ncore/pkg/data/elastic"
	"github.com/ncobase/ncore/pkg/data/meili"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Key constants
const (
	VersionKey      = "version"
	SpanTitleKey    = "title"
	SpanFunctionKey = "function"
)

// Logger represents logger instance
type Logger struct {
	*logrus.Logger
	version     string
	logFile     *os.File
	logPath     string
	meiliClient *meili.Client
	esClient    *elastic.Client
	indexName   string // Meilisearch / Elasticsearch index name
}

var (
	// stdLogger is the global logger
	stdLogger *Logger
	// once ensures that the logger is initialized only once
	once sync.Once
)

// StdLogger returns the single logger instance
func StdLogger() *Logger {
	once.Do(func() {
		stdLogger = &Logger{
			Logger: logrus.New(),
		}
		stdLogger.SetFormatter(&logrus.JSONFormatter{})
	})
	return stdLogger
}

// SetVersion sets the version for logging
func (l *Logger) SetVersion(v string) {
	l.version = v
}

// Init initializes the logger with the given configuration
func (l *Logger) Init(c *config.Logger) (func(), error) {
	l.SetLevel(logrus.Level(c.Level))

	switch c.Format {
	case "json":
		l.SetFormatter(&logrus.JSONFormatter{})
	default:
		l.SetFormatter(&logrus.TextFormatter{})
	}

	switch c.Output {
	case "stdout":
		l.SetOutput(os.Stdout)
	case "stderr":
		l.SetOutput(os.Stderr)
	case "file":
		l.logPath = c.OutputFile
		if l.logPath != "" {
			if err := l.setupLogFile(); err != nil {
				return nil, err
			}
			go l.periodicLogRotation()
		}
	}

	// Initialize MeiliSearch client
	if c.Meilisearch.Host != "" {
		l.meiliClient = meili.NewMeilisearch(c.Meilisearch.Host, c.Meilisearch.APIKey)
		l.indexName = c.IndexName
		l.AddHook(&MeiliSearchHook{
			client: l.meiliClient,
			index:  l.indexName,
		})
	}

	// Initialize Elasticsearch client
	if len(c.Elasticsearch.Addresses) > 0 {
		var err error
		l.esClient, err = elastic.NewClient(c.Elasticsearch.Addresses, c.Elasticsearch.Username, c.Elasticsearch.Password)
		if err != nil {
			return nil, fmt.Errorf("error initializing Elasticsearch client: %w", err)
		}
		l.indexName = c.IndexName
		l.AddHook(&ElasticSearchHook{
			client: l.esClient,
			index:  l.indexName,
		})
	}

	// Return cleanup function
	return func() {
		if l.logFile != nil {
			_ = l.logFile.Close()
		}
	}, nil
}

// setupLogFile sets up the log file
func (l *Logger) setupLogFile() error {
	if err := os.MkdirAll(filepath.Dir(l.logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	return l.rotateLog()
}

// rotateLog rotates the log
func (l *Logger) rotateLog() error {
	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			return fmt.Errorf("failed to close current log file: %w", err)
		}
	}

	logFilePath := fmt.Sprintf("%s.%s.log", strings.TrimSuffix(l.logPath, ".log"), time.Now().Format("2006-01-02"))
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}

	l.logFile = f
	l.SetOutput(l.logFile)
	return nil
}

// periodicLogRotation rotates the log every 24 hours
func (l *Logger) periodicLogRotation() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := l.rotateLog(); err != nil {
			l.Logger.Errorf("Error rotating log: %v", err)
		}
	}
}

// entryFromContext creates a new log entry with fields from context
func (l *Logger) entryFromContext(ctx context.Context) *logrus.Entry {
	fields := logrus.Fields{}

	traceID := getTraceID(ctx)
	if traceID != "" {
		fields[traceKey] = traceID
	}

	if l.version != "" {
		fields[VersionKey] = l.version
	}

	return l.WithFields(fields)
}

// Log methods

// Log logs a message with the given level
func (l *Logger) log(ctx context.Context, level logrus.Level, args ...any) {
	l.entryFromContext(ctx).Log(level, args...)
}

// Logf logs a formatted message
func (l *Logger) logf(ctx context.Context, level logrus.Level, format string, args ...any) {
	l.entryFromContext(ctx).Logf(level, format, args...)
}

// Trace logs a trace message
func (l *Logger) Trace(ctx context.Context, args ...any) {
	l.log(ctx, logrus.TraceLevel, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, args ...any) {
	l.log(ctx, logrus.DebugLevel, args...)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, args ...any) {
	l.log(ctx, logrus.InfoLevel, args...)
}

// Warn logs a warn message
func (l *Logger) Warn(ctx context.Context, args ...any) {
	l.log(ctx, logrus.WarnLevel, args...)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, args ...any) {
	l.log(ctx, logrus.ErrorLevel, args...)
}

// Fatal logs a fatal message
func (l *Logger) Fatal(ctx context.Context, args ...any) {
	l.log(ctx, logrus.FatalLevel, args...)
}

// Panic logs a panic message
func (l *Logger) Panic(ctx context.Context, args ...any) {
	l.log(ctx, logrus.PanicLevel, args...)
}

// Tracef logs a trace message with format
func (l *Logger) Tracef(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.TraceLevel, format, args...)
}

// Debugf logs a debug message with format
func (l *Logger) Debugf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.DebugLevel, format, args...)
}

// Infof logs an info message with format
func (l *Logger) Infof(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.InfoLevel, format, args...)
}

// Warnf logs a warn message with format
func (l *Logger) Warnf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.WarnLevel, format, args...)
}

// Errorf logs an error message with format
func (l *Logger) Errorf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.ErrorLevel, format, args...)
}

// Fatalf logs a fatal message with format
func (l *Logger) Fatalf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.FatalLevel, format, args...)
}

// Panicf logs a panic message with format
func (l *Logger) Panicf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.PanicLevel, format, args...)
}

// MeiliSearch and Elasticsearch log hooks

// MeiliSearchHook represents a MeiliSearch log hook
type MeiliSearchHook struct {
	client *meili.Client
	index  string
}

// Levels returns all log levels
func (h *MeiliSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends log entry to MeiliSearch
func (h *MeiliSearchHook) Fire(entry *logrus.Entry) error {
	jsonData, err := json.Marshal(entry.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal log data: %w", err)
	}
	return h.client.IndexDocuments(h.index, jsonData)
}

// ElasticSearchHook represents an Elasticsearch log hook
type ElasticSearchHook struct {
	client *elastic.Client
	index  string
}

// Levels returns all log levels
func (h *ElasticSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends log entry to Elasticsearch
func (h *ElasticSearchHook) Fire(entry *logrus.Entry) error {
	return h.client.IndexDocument(context.Background(), h.index, entry.Time.Format(time.RFC3339), entry.Data)
}

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(out io.Writer) {
	l.Logger.SetOutput(out)
}

// AddHook adds a hook to the logger
func (l *Logger) AddHook(hook logrus.Hook) {
	if !l.hookExists(hook) {
		l.Logger.AddHook(hook)
	}
}

// hookExists checks if hook already exists
func (l *Logger) hookExists(hook logrus.Hook) bool {
	for _, h := range l.Hooks {
		for _, existingHook := range h {
			if existingHook == hook {
				return true
			}
		}
	}
	return false
}

// SetVersion sets the version for logging
func SetVersion(v string) { StdLogger().SetVersion(v) }

// New creates new logger
func New(c *config.Logger) (func(), error) { return StdLogger().Init(c) }

// WithFields returns an entry with the given fields
func WithFields(ctx context.Context, fields logrus.Fields) *logrus.Entry {
	entry := StdLogger().entryFromContext(ctx)
	return entry.WithFields(fields)
}

// Trace logs trace message
func Trace(ctx context.Context, args ...any) { StdLogger().Trace(ctx, args...) }

// Debug logs debug message
func Debug(ctx context.Context, args ...any) { StdLogger().Debug(ctx, args...) }

// Info logs info message
func Info(ctx context.Context, args ...any) { StdLogger().Info(ctx, args...) }

// Warn logs warn message
func Warn(ctx context.Context, args ...any) { StdLogger().Warn(ctx, args...) }

// Error logs error message
func Error(ctx context.Context, args ...any) { StdLogger().Error(ctx, args...) }

// Fatal logs fatal message
func Fatal(ctx context.Context, args ...any) { StdLogger().Fatal(ctx, args...) }

// Panic logs panic message
func Panic(ctx context.Context, args ...any) { StdLogger().Panic(ctx, args...) }

// Tracef logs trace message with format
func Tracef(ctx context.Context, format string, args ...any) {
	StdLogger().Tracef(ctx, format, args...)
}

// Debugf logs debug message with format
func Debugf(ctx context.Context, format string, args ...any) {
	StdLogger().Debugf(ctx, format, args...)
}

// Infof logs info message with format
func Infof(ctx context.Context, format string, args ...any) {
	StdLogger().Infof(ctx, format, args...)
}

// Warnf logs warn message with format
func Warnf(ctx context.Context, format string, args ...any) {
	StdLogger().Warnf(ctx, format, args...)
}

// Errorf logs error message with format
func Errorf(ctx context.Context, format string, args ...any) {
	StdLogger().Errorf(ctx, format, args...)
}

// Fatalf logs fatal message with format
func Fatalf(ctx context.Context, format string, args ...any) {
	StdLogger().Fatalf(ctx, format, args...)
}

// Panicf logs panic message with format
func Panicf(ctx context.Context, format string, args ...any) {
	StdLogger().Panicf(ctx, format, args...)
}

// SetOutput sets the output destination for the logger
func SetOutput(out io.Writer) { StdLogger().SetOutput(out) }

// AddHook adds a hook to the logger
func AddHook(hook logrus.Hook) { StdLogger().AddHook(hook) }
