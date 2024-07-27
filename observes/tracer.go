package observes

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type TracerOption struct {
	URL                string
	Name               string
	Version            string
	Branch             string
	Revision           string
	Environment        string
	SamplingRate       float64
	MaxAttributes      int
	BatchTimeout       time.Duration
	ExportTimeout      time.Duration
	MaxExportBatchSize int
}

func NewTracer(opt *TracerOption) error {
	if opt == nil {
		return fmt.Errorf("tracer config is nil")
	}

	exp, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(opt.URL),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(opt.Name),
			attribute.String("version", opt.Version),
			attribute.String("branch", opt.Branch),
			attribute.String("revision", opt.Revision),
			attribute.String("environment", opt.Environment),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(opt.SamplingRate))),
		sdktrace.WithBatcher(exp,
			sdktrace.WithMaxExportBatchSize(opt.MaxExportBatchSize),
			sdktrace.WithBatchTimeout(opt.BatchTimeout),
			sdktrace.WithExportTimeout(opt.ExportTimeout),
		),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return nil
}

type Layer int

const (
	LayerUnknown Layer = iota
	LayerHandler
	LayerService
	LayerRepo
)

func (l Layer) String() string {
	return [...]string{"Unknown", "Handler", "Service", "Repository"}[l]
}

type MethodCall struct {
	Layer Layer
	Name  string
}

type TracingContext struct {
	ctx         context.Context
	span        trace.Span
	methodCalls []MethodCall
	maxAttrs    int
}

func NewTracingContext(ctx context.Context, name string, maxAttrs int) *TracingContext {
	tracer := otel.Tracer("")
	ctx, span := tracer.Start(ctx, name)
	return &TracingContext{
		ctx:         ctx,
		span:        span,
		methodCalls: []MethodCall{},
		maxAttrs:    maxAttrs,
	}
}

func (tc *TracingContext) AddMethodCall(layer Layer, name string) {
	tc.methodCalls = append(tc.methodCalls, MethodCall{Layer: layer, Name: name})
}

func (tc *TracingContext) updateSpanAttributes() {
	attrs := make([]attribute.KeyValue, 0, tc.maxAttrs)
	layerCalls := make(map[Layer][]string)

	for _, call := range tc.methodCalls {
		layerCalls[call.Layer] = append(layerCalls[call.Layer], call.Name)
	}

	for layer, calls := range layerCalls {
		if len(attrs) < tc.maxAttrs {
			attrs = append(attrs, attribute.StringSlice(layer.String()+"_calls", calls))
		}
	}

	tc.span.SetAttributes(attrs...)
}

func (tc *TracingContext) SetAttributes(attributes ...attribute.KeyValue) {
	tc.span.SetAttributes(attributes...)
}

func (tc *TracingContext) SetStatus(code codes.Code, description string) {
	tc.span.SetStatus(code, description)
}

func (tc *TracingContext) Context() context.Context {
	return tc.ctx
}

func (tc *TracingContext) End() {
	tc.updateSpanAttributes()
	tc.span.End()
}

type TracingDecoratorOption struct {
	Layer                    Layer
	CreateSpanForEachMethod  bool
	RecordMethodParams       bool
	RecordMethodReturnValues bool
}

func DecorateStruct[T any](obj T, opt TracingDecoratorOption) T {
	return decorateValue(reflect.ValueOf(obj), opt).Interface().(T)
}

func decorateValue(v reflect.Value, opt TracingDecoratorOption) reflect.Value {
	if v.Kind() != reflect.Ptr {
		return v
	}

	elem := v.Elem()
	switch elem.Kind() {
	case reflect.Struct:
		return decorateStruct(v, opt)
	case reflect.Interface:
		return decorateInterface(v, opt)
	default:
		return v
	}
}

func decorateStruct(v reflect.Value, opt TracingDecoratorOption) reflect.Value {
	elemType := v.Elem().Type()
	newElem := reflect.New(elemType).Elem()
	newElem.Set(v.Elem())

	for i := 0; i < elemType.NumField(); i++ {
		field := newElem.Field(i)
		if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
			decoratedField := decorateValue(field, opt)
			if decoratedField.IsValid() {
				field.Set(decoratedField)
			}
		}
	}

	return newElem.Addr()
}

func decorateInterface(v reflect.Value, opt TracingDecoratorOption) reflect.Value {
	if v.IsNil() {
		return v
	}

	elemValue := v.Elem()
	decoratedValue := decorateValue(elemValue, opt)

	if decoratedValue.Type() != elemValue.Type() {
		wrapperType := reflect.StructOf([]reflect.StructField{
			{
				Name: "Original",
				Type: v.Type(),
			},
		})
		wrapper := reflect.New(wrapperType).Elem()
		wrapper.Field(0).Set(decoratedValue)

		for i := 0; i < v.Type().NumMethod(); i++ {
			method := v.Type().Method(i)
			implementMethod(wrapper.Addr(), method, opt)
		}

		return wrapper.Addr()
	}

	return decoratedValue
}

func implementMethod(structValue reflect.Value, method reflect.Method, opt TracingDecoratorOption) {
	newMethod := reflect.MakeFunc(method.Type, func(args []reflect.Value) []reflect.Value {
		methodName := fmt.Sprintf("%s.%s", method.Type.In(0).Elem().Name(), method.Name)

		var span trace.Span
		var ctx context.Context

		if len(args) > 0 && args[0].Type().Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			ctx = args[0].Interface().(context.Context)
		} else {
			ctx = context.Background()
		}

		if opt.CreateSpanForEachMethod {
			ctx, span = otel.Tracer("").Start(ctx, methodName)
			defer span.End()
		}

		tc := getTracingContext(ctx)
		if tc == nil {
			tc = NewTracingContext(ctx, methodName, 100)
			ctx = context.WithValue(ctx, "tracing_context", tc)
			defer tc.End()
		}

		tc.AddMethodCall(opt.Layer, methodName)

		if opt.RecordMethodParams {
			for i, arg := range args {
				if i == 0 && arg.Type().Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
					continue
				}
				span.SetAttributes(attribute.String(fmt.Sprintf("arg%d", i), fmt.Sprintf("%+v", arg.Interface())))
			}
		}

		if len(args) > 0 && args[0].Type().Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			args[0] = reflect.ValueOf(ctx)
		}

		original := structValue.Elem().Field(0)
		results := method.Func.Call(append([]reflect.Value{original}, args...))

		if opt.RecordMethodReturnValues {
			for i, result := range results {
				span.SetAttributes(attribute.String(fmt.Sprintf("result%d", i), fmt.Sprintf("%+v", result.Interface())))
			}
		}

		if len(results) > 0 {
			lastResult := results[len(results)-1].Interface()
			if err, ok := lastResult.(error); ok && err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "Success")
			}
		}

		return results
	})

	structValue.Method(method.Index).Set(newMethod)
}

func getTracingContext(ctx context.Context) *TracingContext {
	if tc, ok := ctx.Value("tracing_context").(*TracingContext); ok {
		return tc
	}
	return nil
}
