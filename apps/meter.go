package main

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.GetMeterProvider().Meter("api")
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

	return func(c echo.Context) error {
		ctx := c.Request().Context()
		path := c.Path()

		slog.InfoContext(ctx, path)

		start := time.Now()
		counter.Add(c.Request().Context(), 1, metric.WithAttributes(attribute.String("api.path", path)))
		if err = next(c); err != nil {
			return err
		}
		duration := time.Since(start)
		histogram.Record(c.Request().Context(), duration.Seconds(), metric.WithAttributes(attribute.String("api.path", path)))
		return nil
	}
}
