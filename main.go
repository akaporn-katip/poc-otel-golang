package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/akapond-katip/poc-traces-metrics-and-logs-golang/api/hello"
	loggerhandler "github.com/akapond-katip/poc-traces-metrics-and-logs-golang/logger-handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	slogecho "github.com/samber/slog-echo"
)

type RequestID string

const (
	TraceIDKey             = "trace_id"
	SpanIDKey              = "span_id"
	RequestIDKey RequestID = "request_id"
)

func main() {
	logHandler := loggerhandler.CreateHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: false,
	}), func(ctx context.Context) []slog.Attr {
		attrs := []slog.Attr{}
		if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
			attrs = append(attrs, slog.String("request_id", requestID))
		}
		return attrs
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	config := slogecho.Config{
		WithRequestBody:    true,
		WithResponseBody:   true,
		WithRequestHeader:  true,
		WithResponseHeader: true,
		WithSpanID:         true,
		WithTraceID:        true,
		WithUserAgent:      true,
	}

	slogecho.TraceIDKey = TraceIDKey
	slogecho.SpanIDKey = SpanIDKey

	e := echo.New()
	e.HideBanner = true
	e.Logger.SetLevel(log.DEBUG)

	// middlewares
	e.Use(middleware.RequestID())
	e.Use(slogecho.NewWithConfig(logger, config))
	e.Use(middleware.Recover())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			ctx := context.WithValue(c.Request().Context(), RequestIDKey, requestID)
			request := c.Request().Clone(ctx)
			c.SetRequest(request)

			logger.InfoContext(c.Request().Context(), "asdf")
			if err := next(c); err != nil {
				return err
			}
			return nil
		}
	})

	// routes
	e.GET("/hello", hello.HelloApiHandler)

	e.Logger.Fatal(e.Start(":3333"))
}
