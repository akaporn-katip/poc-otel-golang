package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/akapond-katip/poc-otel-golang/api/hello"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx := context.Background()
	shutdown, err := setUpOtelSDK(ctx)
	if err != nil {
		slog.Error("failed to set up OpenTelemetry SDK", "error", err)
		os.Exit(1)
	}

	defer func() {
		if err := shutdown(ctx); err != nil {
			slog.Error("failed to shut down OpenTelemetry SDK", "error", err)
			os.Exit(1)
		}
	}()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// middlewares
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware(serviceName))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			span := trace.SpanFromContext(c.Request().Context())
			span.SetAttributes(attribute.String("api.version", "v1"))
			return next(c)
		}
	})

	// routes
	e.GET("/hello", hello.HelloApiHandler)

	e.Logger.Fatal(e.Start(":3333"))
}
