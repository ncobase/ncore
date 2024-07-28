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
	"ncobase/common/uuid"

	"github.com/sirupsen/logrus"
)

// Key constants
const (
	TraceIDKey      = "trace_id"
	UserIDKey       = "user_id"
	VersionKey      = "version"
	SpanTitleKey    = "title"
	SpanFunctionKey = "function"
)

var (
	standardLogger *logrus.Logger
	once           sync.Once
	version        string
	logFile        *os.File
	logPath        string
	meiliClient    *meili.Client
	esClient       *elastic.Client
	indexName      string // Meilisearch / Elasticsearch index name
)

// StandardLogger returns the singleton logger instance
func StandardLogger() *logrus.Logger {
	once.Do(func() {
		standardLogger = logrus.New()
		standardLogger.SetFormatter(&logrus.JSONFormatter{})
	})
	return standardLogger
}

// SetVersion sets the version for logging
func SetVersion(v string) {
	version = v
}

// Init initializes the logger with the given configuration
func Init(c *config.Logger) (func(), error) {
	logger := StandardLogger()
	logger.SetLevel(logrus.Level(c.Level))

	switch c.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
	default:
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	switch c.Output {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "file":
		logPath = c.OutputFile
		if logPath != "" {
			if err := setupLogFile(); err != nil {
				return nil, err
			}
			go periodicLogRotation()
		}
	}

	// // Initialize MeiliSearch client
	// if c.Meilisearch.Host != "" {
	// 	meiliClient = meili.NewMeilisearch(c.Meilisearch.Host, c.Meilisearch.APIKey)
	// 	indexName = c.IndexName
	// 	AddMeiliSearchHook()
	// }
	//
	// // Initialize Elasticsearch client
	// if len(c.Elasticsearch.Addresses) > 0 {
	// 	var err error
	// 	esClient, err = elastic.NewClient(c.Elasticsearch.Addresses, c.Elasticsearch.Username, c.Elasticsearch.Password)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error initializing Elasticsearch client: %w", err)
	// 	}
	// 	indexName = c.IndexName
	// 	AddElasticSearchHook()
	// }

	// Return cleanup function
	return func() {
		if logFile != nil {
			_ = logFile.Close()
		}
	}, nil
}

func setupLogFile() error {
	if err := os.MkdirAll(filepath.Dir(logPath), 0777); err != nil {
		return err
	}
	return rotateLog()
}

func rotateLog() error {
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			return err
		}
	}

	logFilePath := fmt.Sprintf("%s.%s.log", strings.TrimSuffix(logPath, ".log"), time.Now().Format("2006-01-02"))
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	logFile = f
	StandardLogger().SetOutput(logFile)
	return nil
}

func periodicLogRotation() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := rotateLog(); err != nil {
			StandardLogger().Errorf("Error rotating log: %v", err)
		}
	}
}

// ContextWithTraceID ensures a trace ID is present in the context
func ContextWithTraceID(ctx context.Context) context.Context {
	if GetTraceID(ctx) == "" {
		return context.WithValue(ctx, TraceIDKey, uuid.New().String())
	}
	return ctx
}

// GetTraceID retrieves the trace ID from the context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// EntryFromContext creates a new log entry with fields from context
func EntryFromContext(ctx context.Context) *logrus.Entry {
	fields := logrus.Fields{}

	if traceID := GetTraceID(ctx); traceID != "" {
		fields[TraceIDKey] = traceID
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		fields[UserIDKey] = userID
	}

	if version != "" {
		fields[VersionKey] = version
	}

	return StandardLogger().WithFields(fields)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	EntryFromContext(ctx).Debugf(format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	EntryFromContext(ctx).Infof(format, args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	EntryFromContext(ctx).Warnf(format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	EntryFromContext(ctx).Errorf(format, args...)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	EntryFromContext(ctx).Fatalf(format, args...)
}

// MeiliSearch and Elasticsearch log hooks

type MeiliSearchHook struct{}

func (h *MeiliSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *MeiliSearchHook) Fire(entry *logrus.Entry) error {
	if meiliClient == nil {
		return nil
	}
	jsonData, err := json.Marshal(entry.Data)
	if err != nil {
		return err
	}
	return meiliClient.IndexDocuments(indexName, jsonData)
}

type ElasticSearchHook struct{}

func (h *ElasticSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *ElasticSearchHook) Fire(entry *logrus.Entry) error {
	if esClient == nil {
		return nil
	}
	return esClient.IndexDocument(context.Background(), indexName, entry.Time.Format(time.RFC3339), entry.Data)
}

// AddMeiliSearchHook adds MeiliSearch hook to logrus
func AddMeiliSearchHook() {
	if meiliClient != nil {
		StandardLogger().AddHook(&MeiliSearchHook{})
	}
}

// AddElasticSearchHook adds Elasticsearch hook to logrus
func AddElasticSearchHook() {
	if esClient != nil {
		StandardLogger().AddHook(&ElasticSearchHook{})
	}
}

// SetOutput sets the output destination for the logger
func SetOutput(out io.Writer) {
	StandardLogger().SetOutput(out)
}

// AddHook adds a hook to the standard logger
func AddHook(hook logrus.Hook) {
	StandardLogger().AddHook(hook)
}
