package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ncobase/common/config"
	"ncobase/common/elastic"
	"ncobase/common/meili"

	"github.com/sirupsen/logrus"
)

// Key constants
const (
	VersionKey      = "version"
	SpanTitleKey    = "title"
	SpanFunctionKey = "function"
)

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
	standardLogger *Logger
	once           sync.Once
)

// StandardLogger returns the singleton logger instance
func StandardLogger() *Logger {
	once.Do(func() {
		standardLogger = &Logger{
			Logger: logrus.New(),
		}
		standardLogger.SetFormatter(&logrus.JSONFormatter{})
	})
	return standardLogger
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
		l.AddMeiliSearchHook()
	}

	// Initialize Elasticsearch client
	if len(c.Elasticsearch.Addresses) > 0 {
		var err error
		l.esClient, err = elastic.NewClient(c.Elasticsearch.Addresses, c.Elasticsearch.Username, c.Elasticsearch.Password)
		if err != nil {
			return nil, fmt.Errorf("error initializing Elasticsearch client: %w", err)
		}
		l.indexName = c.IndexName
		l.AddElasticSearchHook()
	}

	// Return cleanup function
	return func() {
		if l.logFile != nil {
			_ = l.logFile.Close()
		}
	}, nil
}

func (l *Logger) setupLogFile() error {
	if err := os.MkdirAll(filepath.Dir(l.logPath), 0777); err != nil {
		return err
	}
	return l.rotateLog()
}

func (l *Logger) rotateLog() error {
	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			return err
		}
	}

	logFilePath := fmt.Sprintf("%s.%s.log", strings.TrimSuffix(l.logPath, ".log"), time.Now().Format("2006-01-02"))
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	l.logFile = f
	l.SetOutput(l.logFile)
	return nil
}

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
func (l *Logger) log(ctx context.Context, level logrus.Level, args ...any) {
	l.entryFromContext(ctx).Log(level, args...)
}

func (l *Logger) logf(ctx context.Context, level logrus.Level, format string, args ...any) {
	l.entryFromContext(ctx).Logf(level, format, args...)
}

func (l *Logger) Trace(ctx context.Context, args ...any) {
	l.log(ctx, logrus.TraceLevel, args...)
}
func (l *Logger) Debug(ctx context.Context, args ...any) {
	l.log(ctx, logrus.DebugLevel, args...)
}
func (l *Logger) Info(ctx context.Context, args ...any) {
	l.log(ctx, logrus.InfoLevel, args...)
}
func (l *Logger) Warn(ctx context.Context, args ...any) {
	l.log(ctx, logrus.WarnLevel, args...)
}
func (l *Logger) Error(ctx context.Context, args ...any) {
	l.log(ctx, logrus.ErrorLevel, args...)
}
func (l *Logger) Fatal(ctx context.Context, args ...any) {
	l.log(ctx, logrus.FatalLevel, args...)
}
func (l *Logger) Panic(ctx context.Context, args ...any) {
	l.log(ctx, logrus.PanicLevel, args...)
}

func (l *Logger) Tracef(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.TraceLevel, format, args...)
}
func (l *Logger) Debugf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.DebugLevel, format, args...)
}
func (l *Logger) Infof(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.InfoLevel, format, args...)
}
func (l *Logger) Warnf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.WarnLevel, format, args...)
}
func (l *Logger) Errorf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.ErrorLevel, format, args...)
}
func (l *Logger) Fatalf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.FatalLevel, format, args...)
}
func (l *Logger) Panicf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, logrus.PanicLevel, format, args...)
}

// MeiliSearch and Elasticsearch log hooks

type MeiliSearchHook struct {
	client *meili.Client
	index  string
}

func (h *MeiliSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *MeiliSearchHook) Fire(entry *logrus.Entry) error {
	jsonData, err := json.Marshal(entry.Data)
	if err != nil {
		return err
	}
	return h.client.IndexDocuments(h.index, jsonData)
}

type ElasticSearchHook struct {
	client *elastic.Client
	index  string
}

func (h *ElasticSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *ElasticSearchHook) Fire(entry *logrus.Entry) error {
	return h.client.IndexDocument(context.Background(), h.index, entry.Time.Format(time.RFC3339), entry.Data)
}

// AddMeiliSearchHook adds MeiliSearch hook to logrus
func (l *Logger) AddMeiliSearchHook() {
	if l.meiliClient != nil {
		hook := &MeiliSearchHook{
			client: l.meiliClient,
			index:  l.indexName,
		}
		if !l.hookExists(hook) {
			l.AddHook(hook)
		}
	}
}

// AddElasticSearchHook adds Elasticsearch hook to logrus
func (l *Logger) AddElasticSearchHook() {
	if l.esClient != nil {
		hook := &ElasticSearchHook{
			client: l.esClient,
			index:  l.indexName,
		}
		if !l.hookExists(hook) {
			l.AddHook(hook)
		}
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

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(out io.Writer) {
	l.Logger.SetOutput(out)
}

// AddHook adds a hook to the logger
func (l *Logger) AddHook(hook logrus.Hook) {
	l.Logger.AddHook(hook)
}

// Exported functions for backward compatibility

func SetVersion(v string)                   { StandardLogger().SetVersion(v) }
func Init(c *config.Logger) (func(), error) { return StandardLogger().Init(c) }

func EntryWithFields(ctx context.Context, fields logrus.Fields) *logrus.Entry {
	entry := StandardLogger().entryFromContext(ctx)
	return entry.WithFields(fields)
}

func Trace(ctx context.Context, args ...any) { StandardLogger().Trace(ctx, args...) }
func Debug(ctx context.Context, args ...any) { StandardLogger().Debug(ctx, args...) }
func Info(ctx context.Context, args ...any)  { StandardLogger().Info(ctx, args...) }
func Warn(ctx context.Context, args ...any)  { StandardLogger().Warn(ctx, args...) }
func Error(ctx context.Context, args ...any) { StandardLogger().Error(ctx, args...) }
func Fatal(ctx context.Context, args ...any) { StandardLogger().Fatal(ctx, args...) }
func Panic(ctx context.Context, args ...any) { StandardLogger().Panic(ctx, args...) }

func Tracef(ctx context.Context, format string, args ...any) {
	StandardLogger().Tracef(ctx, format, args...)
}
func Debugf(ctx context.Context, format string, args ...any) {
	StandardLogger().Debugf(ctx, format, args...)
}
func Infof(ctx context.Context, format string, args ...any) {
	StandardLogger().Infof(ctx, format, args...)
}
func Warnf(ctx context.Context, format string, args ...any) {
	StandardLogger().Warnf(ctx, format, args...)
}
func Errorf(ctx context.Context, format string, args ...any) {
	StandardLogger().Errorf(ctx, format, args...)
}
func Fatalf(ctx context.Context, format string, args ...any) {
	StandardLogger().Fatalf(ctx, format, args...)
}
func Panicf(ctx context.Context, format string, args ...any) {
	StandardLogger().Panicf(ctx, format, args...)
}

func SetOutput(out io.Writer)  { StandardLogger().SetOutput(out) }
func AddHook(hook logrus.Hook) { StandardLogger().AddHook(hook) }
