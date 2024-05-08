package bootstrap

import (
	"context"
	"ecst-order/internal/appctx"
	"ecst-order/pkg/logger"
	"fmt"
	"go.opentelemetry.io/otel/trace/noop"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// RegistryOpenTelemetry setup
func RegistryOpenTelemetry(cfg *appctx.Config) func() {
	if !cfg.APM.Enable {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return func() {}
	}

	ctx := context.Background()
	lf := logger.NewFields(logger.EventName("TracerInitiated"))
	logger.Debug(fmt.Sprint("apm address : ", cfg.APM.Address), lf...)

	// setup trace resource provider
	rsc, err := resource.New(context.Background(),
		resource.WithFromEnv(),
		resource.WithContainer(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			// service name used to display traces in backends
			semconv.ServiceNameKey.String(cfg.APM.Name),
			// optional
			attribute.String("library.lang", "go"),
			attribute.String("service.env", cfg.App.Env),
			attribute.String("service.address", cfg.APM.Address),
		),
	)
	if err != nil {
		log.Fatalf("failed to create collector resource: %v", err)
	}

	// setup a trace exporter
	traceClient := otlptracehttp.NewClient(
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(cfg.APM.Address),
	)

	exporter, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		log.Fatalf("failed to create trace exporter: %v", err)
	}

	// setup tracer provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		sdktrace.WithResource(rsc),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func() {
		if err = tracerProvider.Shutdown(ctx); err != nil {
			log.Fatalf("failed to stop exporter: %v", err)
		}

		logger.InfoWithContext(ctx, "The tracer provider done stopped")
	}
}
