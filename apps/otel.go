package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	loggrpc "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	metergrpc "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	tracegrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

const (
	serviceName    = "github.com/akapond-katip/poc-otel-golang"
	serviceVersion = "1.0.0"
	environment    = "development"
)

func setUpOtelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	res, err := newResources()
	if err != nil {
		handleErr(err)
		return nil, err
	}

	// tracer
	traceProvider, err := newTraceProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, traceProvider.Shutdown)
	otel.SetTracerProvider(traceProvider)

	// logger
	loggerProvider, err := newLoggerProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)
	slog.SetDefault(otelslog.NewLogger(serviceName))

	// metrics
	meterProvider, err := newMetricsProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)
	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newLogExporter(ctx context.Context) (exp otellog.Exporter, err error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "grpc" {
		exp, err = loggrpc.New(ctx, loggrpc.WithInsecure())
	} else {
		exp, err = stdoutlog.New()
	}

	return
}

func newLoggerProvider(ctx context.Context, res *resource.Resource) (*otellog.LoggerProvider, error) {
	exp, err := newLogExporter(ctx)
	if err != nil {
		return nil, err
	}

	logProcessor := otellog.NewBatchProcessor(exp)
	loggerProvider := otellog.NewLoggerProvider(
		otellog.WithProcessor(logProcessor),
		otellog.WithResource(res),
	)
	return loggerProvider, nil
}

func newTraceExporter(ctx context.Context) (exp trace.SpanExporter, err error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "grpc" {
		exp, err = tracegrpc.New(ctx, tracegrpc.WithInsecure())
	} else {
		exp, err = stdouttrace.New()
	}

	return
}

func newTraceProvider(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
	exp, err := newTraceExporter(ctx)
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
		trace.WithSampler(
			trace.ParentBased(trace.TraceIDRatioBased(1)),
		),
	)

	return traceProvider, nil
}

func newMeterExporter(ctx context.Context) (exp metric.Exporter, err error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "grpc" {
		exp, err = metergrpc.New(ctx, metergrpc.WithInsecure())
	} else {
		exp, err = stdoutmetric.New()
	}

	return
}

func newMetricsProvider(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
	exp, err := newMeterExporter(ctx)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exp)),
	)

	return meterProvider, nil
}

func newResources() (res *resource.Resource, err error) {
	res, err = resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironmentName(environment),
		),
	)

	if err != nil {
		return
	}
	return
}
