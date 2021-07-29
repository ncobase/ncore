package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"stone/common/conf"

	"github.com/sirupsen/logrus"
)

// 定义键名
const (
	TraceIDKey      = "trace"
	UserIDKey       = "user"
	SpanTitleKey    = "title"
	SpanFunctionKey = "function"
	VersionKey      = "version"
	StackKey        = "stack"
)

// TraceIDFunc 定义获取跟踪ID的函数
type TraceIDFunc func() string

var (
	version     string
	traceIDFunc TraceIDFunc
	_           = os.Getpid()
)

func init() {
	traceIDFunc = func() string {
		return fmt.Sprintf("%d",
			os.Getpid())
		// time.Now().Format("2006-01-02 15:04:05"))
	}
}

// Logger 定义日志别名
type Logger = logrus.Logger

// Hook 定义日志钩子别名
type Hook = logrus.Hook

// StandardLogger 获取标准日志
func StandardLogger() *Logger {
	return logrus.StandardLogger()
}

// SetLevel设定日志级别
func setLevel(level int) {
	logrus.SetLevel(logrus.Level(level))
}

// SetFormatter 设定日志输出格式
func setFormatter(format string) {
	switch format {
	case "json":
		logrus.SetFormatter(new(logrus.JSONFormatter))
	default:
		logrus.SetFormatter(new(logrus.TextFormatter))
	}
}

// setOutput 设定日志输出
func setOutput(out io.Writer) {
	logrus.SetOutput(out)
}

// SetVersion 设定版本
func SetVersion(v string) {
	version = v
}

// SetTraceIDFunc 设定追踪ID的处理函数
func SetTraceIDFunc(fn TraceIDFunc) {
	traceIDFunc = fn
}

// AddHook 增加日志钩子
func AddHook(hook Hook) {
	logrus.AddHook(hook)
}

type (
	traceIDKey struct{}
	userIDKey  struct{}
)

// NewTraceIDContext 创建跟踪ID上下文
func NewTraceIDContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// FromTraceIDContext 从上下文中获取跟踪ID
func FromTraceIDContext(ctx context.Context) string {
	v := ctx.Value(traceIDKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return traceIDFunc()
}

// NewUserIDContext 创建用户ID上下文
func NewUserIDContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// FromUserIDContext 从上下文中获取用户ID
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

// SpanOption 定义跟踪单元的数据项
type SpanOption func(*spanOptions)

// SetSpanTitle 设置跟踪单元的标题
func SetSpanTitle(title string) SpanOption {
	return func(o *spanOptions) {
		o.Title = title
	}
}

// SetSpanFuncName 设置跟踪单元的函数名
func SetSpanFuncName(funcName string) SpanOption {
	return func(o *spanOptions) {
		o.FuncName = funcName
	}
}

// StartSpan 开始一个追踪单元
func StartSpan(ctx context.Context, opts ...SpanOption) *Entry {
	if ctx == nil {
		ctx = context.Background()
	}
	var o spanOptions
	for _, opt := range opts {
		opt(&o)
	}
	fields := map[string]interface{}{
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

// Debugf 写入调试日志
func Debugf(ctx context.Context, format string, args ...interface{}) {
	StartSpan(ctx).Debugf(format, args...)
}

// Infof 写入消息日志
func Infof(ctx context.Context, format string, args ...interface{}) {
	StartSpan(ctx).Infof(format, args...)
}

// Printf 写入消息日志
func Printf(ctx context.Context, format string, args ...interface{}) {
	StartSpan(ctx).Printf(format, args...)
}

// Warnf 写入警告日志
func Warnf(ctx context.Context, format string, args ...interface{}) {
	StartSpan(ctx).Warnf(format, args...)
}

// Errorf 写入错误日志
func Errorf(ctx context.Context, format string, args ...interface{}) {
	StartSpan(ctx).Errorf(format, args...)
}

// Fatalf 写入重大错误日志
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	StartSpan(ctx).Fatalf(format, args...)
}

// ErrorStack 输出错误栈
func ErrorStack(ctx context.Context, err error) {
	StartSpan(ctx).WithField(StackKey, fmt.Sprintf("%+v", err)).Errorf(err.Error())
}

// Entry 定义统一的日志写入方式
type Entry struct {
	entry *logrus.Entry
}

func newEntry(entry *logrus.Entry) *Entry {
	return &Entry{entry: entry}
}

func (e *Entry) checkAndDelete(fields map[string]interface{}, keys ...string) *Entry {
	for _, key := range keys {
		_, ok := fields[key]
		if ok {
			delete(fields, key)
		}
	}
	return newEntry(e.entry.WithFields(fields))
}

// WithFields 结构化字段写入
func (e *Entry) WithFields(fields map[string]interface{}) *Entry {
	e.checkAndDelete(fields,
		TraceIDKey,
		SpanTitleKey,
		SpanFunctionKey,
		VersionKey)
	return newEntry(e.entry.WithFields(fields))
}

// WithField 结构化字段写入
func (e *Entry) WithField(key string, value interface{}) *Entry {
	return e.WithFields(map[string]interface{}{key: value})
}

// Fatalf 重大错误日志
func (e *Entry) Fatalf(format string, args ...interface{}) {
	e.entry.Fatalf(format, args...)
}

// Errorf 错误日志
func (e *Entry) Errorf(format string, args ...interface{}) {
	e.entry.Errorf(format, args...)
}

// Warnf 警告日志
func (e *Entry) Warnf(format string, args ...interface{}) {
	e.entry.Warnf(format, args...)
}

// Infof 消息日志
func (e *Entry) Infof(format string, args ...interface{}) {
	e.entry.Infof(format, args...)
}

// Printf 消息日志
func (e *Entry) Printf(format string, args ...interface{}) {
	e.entry.Printf(format, args...)
}

// Debugf 写入调试日志
func (e *Entry) Debugf(format string, args ...interface{}) {
	e.entry.Debugf(format, args...)
}

// Init initialize log
func Init(c conf.LoggerConfig) (clean func(), err error) {
	setLevel(c.Level)
	setFormatter(c.Format)
	// 设定日志输出
	var file *os.File
	if c.Output != "" {
		switch c.Output {
		case "stdout":
			setOutput(os.Stdout)
		case "stderr":
			setOutput(os.Stderr)
		case "file":
			if name := c.OutputFile; name != "" {
				_ = os.MkdirAll(filepath.Dir(name), 0777)
				f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					return nil, err
				}
				setOutput(f)
				file = f
			}
		}
	}

	// file close
	clean = func() {
		if file != nil {
			file.Close()
		}
	}

	return
}
