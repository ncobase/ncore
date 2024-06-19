package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ncobase/common/config"
	"github.com/ncobase/common/elastic"
	"github.com/ncobase/common/meili"

	"github.com/sirupsen/logrus"
)

// Define keys
const (
	TraceIDKey      = "trace_id"
	UserIDKey       = "user_id"
	SpanTitleKey    = "title"
	SpanFunctionKey = "function"
	VersionKey      = "version"
	StackKey        = "stack"
)

// TraceIDFunc defines a function to get trace ID.
type TraceIDFunc func() string

var (
	version     string
	traceIDFunc TraceIDFunc
	logFile     *os.File
	logPath     string
	meiliClient *meili.Client
	esClient    *elastic.Client
	indexName   string // Meilisearch / Elasticsearch index name
)

func init() {
	traceIDFunc = func() string {
		return fmt.Sprintf("%d", os.Getpid())
	}
}

// Logger defines an alias for logrus.Logger.
type Logger = logrus.Logger

// Hook defines an alias for logrus.Hook.
type Hook = logrus.Hook

// StandardLogger gets the standard logger.
func StandardLogger() *Logger {
	return logrus.StandardLogger()
}

// SetLevel sets the log level.
func setLevel(level int) {
	logrus.SetLevel(logrus.Level(level))
}

// SetFormatter sets the log formatter.
func setFormatter(format string) {
	switch format {
	case "json":
		logrus.SetFormatter(new(logrus.JSONFormatter))
	default:
		logrus.SetFormatter(new(logrus.TextFormatter))
	}
}

// setOutput sets the log output.
func setOutput(out io.Writer) {
	logrus.SetOutput(out)
}

// SetVersion sets the version.
func SetVersion(v string) {
	version = v
}

// SetTraceIDFunc sets the function to handle trace ID.
func SetTraceIDFunc(fn TraceIDFunc) {
	traceIDFunc = fn
}

// AddHook adds a log hook.
func AddHook(hook Hook) {
	logrus.AddHook(hook)
}

type (
	traceIDKey struct{}
	userIDKey  struct{}
)

// NewTraceIDContext creates a trace ID context.
func NewTraceIDContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// FromTraceIDContext gets the trace ID from context.
func FromTraceIDContext(ctx context.Context) string {
	v := ctx.Value(traceIDKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return traceIDFunc()
}

// NewUserIDContext creates a user ID context.
func NewUserIDContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// FromUserIDContext gets the user ID from context.
func FromUserIDContext(ctx context.Context) string {
	v := ctx.Value(userIDKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

type spanOptions struct {
	Title    string
	FuncName string
}

// SpanOption defines options for a span.
type SpanOption func(*spanOptions)

// SetSpanTitle sets the title for a span.
func SetSpanTitle(title string) SpanOption {
	return func(o *spanOptions) {
		o.Title = title
	}
}

// SetSpanFuncName sets the function name for a span.
func SetSpanFuncName(funcName string) SpanOption {
	return func(o *spanOptions) {
		o.FuncName = funcName
	}
}

// StartSpan starts a new span.
func StartSpan(ctx context.Context, opts ...SpanOption) *Entry {
	if ctx == nil {
		ctx = context.Background()
	}
	var o spanOptions
	for _, opt := range opts {
		opt(&o)
	}
	fields := map[string]any{
		VersionKey: version,
	}
	if v := FromTraceIDContext(ctx); v != "" {
		fields[TraceIDKey] = v
	}
	if v := FromUserIDContext(ctx); v != "" {
		fields[UserIDKey] = v
	}
	if v := o.Title; v != "" {
		fields[SpanTitleKey] = v
	}
	if v := o.FuncName; v != "" {
		fields[SpanFunctionKey] = v
	}

	return newEntry(logrus.WithFields(fields))
}

// Debugf writes debug log.
func Debugf(ctx context.Context, format string, args ...any) {
	StartSpan(ctx).Debugf(format, args...)
}

// Infof writes info log.
func Infof(ctx context.Context, format string, args ...any) {
	StartSpan(ctx).Infof(format, args...)
}

// Printf writes info log.
func Printf(ctx context.Context, format string, args ...any) {
	StartSpan(ctx).Printf(format, args...)
}

// Warnf writes warning log.
func Warnf(ctx context.Context, format string, args ...any) {
	StartSpan(ctx).Warnf(format, args...)
}

// Errorf writes error log.
func Errorf(ctx context.Context, format string, args ...any) {
	StartSpan(ctx).Errorf(format, args...)
}

// Fatalf writes fatal error log.
func Fatalf(ctx context.Context, format string, args ...any) {
	StartSpan(ctx).Fatalf(format, args...)
}

// ErrorStack outputs error stack.
func ErrorStack(ctx context.Context, err error) {
	StartSpan(ctx).WithField(StackKey, fmt.Sprintf("%+v", err)).Errorf(err.Error())
}

// Entry defines a unified way of writing logs.
type Entry struct {
	entry *logrus.Entry
}

func newEntry(entry *logrus.Entry) *Entry {
	return &Entry{entry: entry}
}

func (e *Entry) checkAndDelete(fields map[string]any, keys ...string) *Entry {
	for _, key := range keys {
		delete(fields, key)
	}
	return newEntry(e.entry.WithFields(fields))
}

// WithFields writes structured fields.
func (e *Entry) WithFields(fields map[string]any) *Entry {
	e.checkAndDelete(fields, TraceIDKey, SpanTitleKey, SpanFunctionKey, VersionKey)
	return newEntry(e.entry.WithFields(fields))
}

// WithField writes a structured field.
func (e *Entry) WithField(key string, value any) *Entry {
	return e.WithFields(map[string]any{key: value})
}

// Fatalf writes a fatal error log.
func (e *Entry) Fatalf(format string, args ...any) {
	e.entry.Fatalf(format, args...)
}

// Errorf writes an error log.
func (e *Entry) Errorf(format string, args ...any) {
	e.entry.Errorf(format, args...)
}

// Warnf writes a warning log.
func (e *Entry) Warnf(format string, args ...any) {
	e.entry.Warnf(format, args...)
}

// Infof writes an info log.
func (e *Entry) Infof(format string, args ...any) {
	e.entry.Infof(format, args...)
}

// Printf writes an info log.
func (e *Entry) Printf(format string, args ...any) {
	e.entry.Printf(format, args...)
}

// Debugf writes a debug log.
func (e *Entry) Debugf(format string, args ...any) {
	e.entry.Debugf(format, args...)
}

// Init initializes the log configuration.
func Init(c config.Logger) (clean func(), err error) {
	setLevel(c.Level)
	setFormatter(c.Format)

	// Set log output
	switch c.Output {
	case "stdout":
		setOutput(os.Stdout)
	case "stderr":
		setOutput(os.Stderr)
	case "file":
		logPath = c.OutputFile
		if logPath != "" {
			logDir := filepath.Dir(logPath)
			if logDir != "." { // If logPath contains directory information
				if err := os.MkdirAll(logDir, 0777); err != nil {
					return nil, err
				}
			}
			if err := rotateLog(); err != nil {
				return nil, err
			}
			go periodicLogRotation()
		}
	}

	// // Initialize MeiliSearch and Elasticsearch clients
	// if c.Meilisearch.Host != "" {
	// 	meiliClient = meili.NewMeilisearch(c.Meilisearch.Host, c.Meilisearch.APIKey)
	// 	indexName = c.IndexName
	// 	AddMeiliSearchHook()
	// }
	//
	// if len(c.Elasticsearch.Addresses) > 0 {
	// 	esClient, err = elastic.NewClient(c.Elasticsearch.Addresses, c.Elasticsearch.Username, c.Elasticsearch.Password)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error initializing Elasticsearch client: %w", err)
	// 	}
	// 	indexName = c.IndexName
	// 	AddElasticSearchHook()
	// }

	// Return a cleanup function to close the file
	clean = func() {
		if logFile != nil {
			_ = logFile.Close()
		}
	}

	return clean, nil
}

func rotateLog() error {
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			return err
		}
	}

	logFilePath := logPath
	if strings.HasSuffix(logPath, ".log") {
		logFilePath = strings.TrimSuffix(logPath, ".log") + "." + time.Now().Format(time.DateOnly) + ".log"
	} else {
		logFilePath = logPath + "." + time.Now().Format(time.DateOnly) + ".log"
	}
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	logFile = f
	setOutput(logFile)
	return nil
}

func periodicLogRotation() {
	// Rotate logs every day at midnight
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := rotateLog(); err != nil {
				Errorf(context.Background(), "Error rotating log: %v", err)
			}
		}
	}
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
	err = meiliClient.IndexDocuments(indexName, jsonData)
	return err
}

type ElasticSearchHook struct{}

func (h *ElasticSearchHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *ElasticSearchHook) Fire(entry *logrus.Entry) error {
	if esClient == nil {
		return nil
	}
	err := esClient.IndexDocument(context.Background(), indexName, entry.Time.Format(time.RFC3339), entry.Data)
	return err
}

// AddMeiliSearchHook Adding MeiliSearch hook to logrus
func AddMeiliSearchHook() {
	if meiliClient != nil {
		AddHook(&MeiliSearchHook{})
	}
}

// AddElasticSearchHook Adding Elasticsearch hook to logrus
func AddElasticSearchHook() {
	if esClient != nil {
		AddHook(&ElasticSearchHook{})
	}
}
