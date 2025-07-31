package main

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracegrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

const (
	serviceName    = "github.com/akapond-katip/poc-traces-metrics-and-logs-golang"
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

	grpcExp, err := newExporter(ctx)
	if err != nil {
		handleErr(err)
		return nil, err
	}

	// tracer
	traceProvider, err := newTraceProvider(grpcExp)
	if err != nil {
		handleErr(err)
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, traceProvider.Shutdown)
	otel.SetTracerProvider(traceProvider)

	// logger
	loggerProvider, err := newLoggerProvider()
	if err != nil {
		handleErr(err)
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)
	slog.SetDefault(otelslog.NewLogger(serviceName))
	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newLoggerProvider() (*otellog.LoggerProvider, error) {
	exp, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	logProcessor := otellog.NewBatchProcessor(exp)
	loggerProvider := otellog.NewLoggerProvider(otellog.WithProcessor(logProcessor))
	return loggerProvider, nil
}

func newExporter(ctx context.Context) (exp *otlptrace.Exporter, err error) {
	exp, err = tracegrpc.New(ctx, tracegrpc.WithInsecure())
	if err != nil {
		return
	}

	return
}

func newTraceProvider(exporter trace.SpanExporter) (*trace.TracerProvider, error) {
	rs, err := newResources()
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(rs),
		trace.WithSampler(
			trace.ParentBased(trace.TraceIDRatioBased(1)),
		),
	)

	return traceProvider, nil
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
