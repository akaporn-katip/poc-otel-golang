package main

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.GetMeterProvider().Meter("echo")
)

func MeterMiddleware(next echo.HandlerFunc) echo.HandlerFunc {

	counter, err := meter.Int64Counter(
		"api.counter",
		metric.WithDescription("Number of API calls."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		panic(err)
	}

	histogram, err := meter.Float64Histogram(
		"api.duration",
		metric.WithDescription("The duration of api execution."),
		metric.WithUnit("s"),
	)

	if err != nil {
		panic(err)
	}

	errorCounter, err := meter.Int64Counter(
		"api.error_count",
		metric.WithDescription("Number of API errors."),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		panic(err)
	}

	reqSizeHistogram, err := meter.Int64Histogram(
		"api.request_size",
		metric.WithDescription("Size of API requests in bytes."),
		metric.WithUnit("By"),
	)
	if err != nil {
		panic(err)
	}

	respSizeHistogram, err := meter.Int64Histogram(
		"api.response_size",
		metric.WithDescription("Size of API responses in bytes."),
		metric.WithUnit("By"),
	)
	if err != nil {
		panic(err)
	}

	statusCodeCounter, err := meter.Int64Counter(
		"api.status_code_count",
		metric.WithDescription("Count of API responses by status code."),
		metric.WithUnit("{response}"),
	)
	if err != nil {
		panic(err)
	}

	return func(c echo.Context) error {
		ctx := c.Request().Context()
		path := c.Path()
		method := c.Request().Method

		start := time.Now()
		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("api.path", path),
			attribute.String("api.method", method),
		))

		reqSize := c.Request().ContentLength
		reqSizeHistogram.Record(ctx, int64(reqSize), metric.WithAttributes(
			attribute.String("api.path", path),
			attribute.String("api.method", method),
		))

		err = next(c)
		duration := time.Since(start)
		histogram.Record(ctx, duration.Seconds(), metric.WithAttributes(
			attribute.String("api.path", path),
			attribute.String("api.method", method),
		))

		status := c.Response().Status
		statusCodeCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("api.path", path),
			attribute.String("api.method", method),
			attribute.Int("api.status_code", status),
		))

		respSize := c.Response().Size
		respSizeHistogram.Record(ctx, int64(respSize), metric.WithAttributes(
			attribute.String("api.path", path),
			attribute.String("api.method", method),
		))

		if err != nil || status >= 400 {
			errorCounter.Add(ctx, 1, metric.WithAttributes(
				attribute.String("api.path", path),
				attribute.String("api.method", method),
				attribute.Int("api.status_code", status),
			))
		}
		if err != nil {
			return err
		}
		return nil
	}
}
