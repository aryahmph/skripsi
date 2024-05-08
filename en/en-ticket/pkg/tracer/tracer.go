package tracer

import (
	"context"
	"en-ticket/pkg/util"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Option struct {
	TagKey string
	Value  string
}

func WithOptions(key, value string) Option {
	return Option{
		TagKey: key,
		Value:  value,
	}
}

var tracer trace.Tracer

func WithResourceNameOptions(value string) Option {
	return Option{
		TagKey: "resource.name",
		Value:  value,
	}
}

func init() {
	tracer = NewTracer()
}

func NewTracer() trace.Tracer {
	return otel.Tracer("en-ticket")
}

// SpanStart starts a new query span from ctx, then returns a new context with the new span.
func SpanStart(ctx context.Context, spanName string) context.Context {
	ctx, _ = tracer.Start(ctx, spanName)
	return ctx
}

// SpanStartWithSpan starts a new query span from ctx, then returns a new context with the new span.
func SpanStartWithSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	ctx, sp := tracer.Start(ctx, spanName)
	return ctx, sp
}

// AddEvent add new event in existing trace from ctx
func AddEvent(ctx context.Context, eventName string) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(eventName)
}

// SpanFinish finishes the span associated with ctx.
func SpanFinish(ctx context.Context) {
	if span := trace.SpanFromContext(ctx); span != nil {
		span.End()
	}
}

// SpanError adds an error to the span associated with ctx.
func SpanError(ctx context.Context, err error) {
	if span := trace.SpanFromContext(ctx); span != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error")
	}
}

func SpanStartWithOption(ctx context.Context, eventName string, opts ...Option) context.Context {

	var spOpts []trace.SpanStartOption

	for x := 0; x < len(opts); x++ {
		if util.InArray(opts[x].TagKey, []string{
			"span.type",
			"service.name",
			"resource.name",
		}) {
			spOpts = append(spOpts, trace.WithAttributes(attribute.KeyValue{
				Key:   attribute.Key(opts[x].TagKey),
				Value: attribute.StringValue(opts[x].Value),
			}))
		}
	}

	ctx, sp := tracer.Start(ctx, eventName, spOpts...)

	for i := 0; i < len(opts); i++ {
		sp.SetAttributes(attribute.KeyValue{
			Key:   attribute.Key(opts[i].TagKey),
			Value: attribute.StringValue(opts[i].Value),
		})
	}

	return ctx
}

func SpanStartWithSpanOption(ctx context.Context, eventName string, opts ...Option) (context.Context, trace.Span) {

	var spOpts []trace.SpanStartOption

	for x := 0; x < len(opts); x++ {
		if util.InArray(opts[x].TagKey, []string{
			"span.type",
			"service.name",
			"resource.name",
		}) {
			// spOptions = append(spOptions, opentracing.Tag{Key: opts[x].TagKey, Value: opts[x].Value})
			spOpts = append(spOpts, trace.WithAttributes(attribute.KeyValue{
				Key:   attribute.Key(opts[x].TagKey),
				Value: attribute.StringValue(opts[x].Value),
			}))
		}
	}

	ctx, sp := tracer.Start(ctx, eventName, spOpts...)

	for i := 0; i < len(opts); i++ {
		sp.SetAttributes(attribute.KeyValue{
			Key:   attribute.Key(opts[i].TagKey),
			Value: attribute.StringValue(opts[i].Value),
		})
	}

	return ctx, sp
}

func DBSpanStartWithOption(ctx context.Context, dbName, eventName string, opts ...Option) context.Context {

	svcName := fmt.Sprintf("%s.%s", "mysql", dbName)
	opts = append(opts,
		WithOptions("db.type", "sql"),
		WithOptions("span.kind", "client"),
		WithOptions("peer.service", svcName),
		WithOptions("service.name", svcName),
	)

	return SpanStartWithOption(ctx, eventName, opts...)
}

func KafkaSpanStartWithOption(ctx context.Context, eventName string, opts ...Option) context.Context {

	svcName := "kafka"
	opts = append(opts,
		WithOptions("span.type", "queue"),
		WithOptions("span.kind", "client"),
		WithOptions("peer.service", svcName),
		WithOptions("service.name", svcName),
	)

	return SpanStartWithOption(ctx, eventName, opts...)
}

func InternalAPISpanStartWithOption(ctx context.Context, internalServiceName, eventName string, opts ...Option) context.Context {
	svcName := fmt.Sprintf("%s.%s", "internal", internalServiceName)
	opts = append(opts,
		WithOptions("span.type", "web"),
		WithOptions("span.kind", "client"),
		WithOptions("peer.service", svcName),
		WithOptions("service.name", svcName),
	)

	return SpanStartWithOption(ctx, eventName, opts...)
}
