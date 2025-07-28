package loggerhandler

import (
	"context"
	"log/slog"
)

type Loggerhandler struct {
	slog.Handler
	customFn CustomAttrWithContext
}

type CustomAttrWithContext func(ctx context.Context) []slog.Attr

func (lh Loggerhandler) Enabled(ctx context.Context, level slog.Level) bool {
	return lh.Handler.Enabled(ctx, level)
}

func (lh Loggerhandler) Handle(ctx context.Context, record slog.Record) error {
	// record.AddAttrs(slog.String("request_id", "fake_request_id"))
	record.AddAttrs(lh.customFn(ctx)...)
	return lh.Handler.Handle(ctx, record)
}

func (lh Loggerhandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return lh.Handler.WithAttrs(attrs)
}

func (lh Loggerhandler) WithGroup(name string) slog.Handler {
	return lh.Handler.WithGroup(name)
}

func CreateHandler(handler slog.Handler, customFn CustomAttrWithContext) slog.Handler {
	return &Loggerhandler{
		Handler:  handler,
		customFn: customFn,
	}
}
