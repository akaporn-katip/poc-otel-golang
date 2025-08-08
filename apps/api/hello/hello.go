package hello

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
)

const name = "github.com/akapond-katip/poc-traces-metrics-and-logs-golang"

var (
	tracer = otel.Tracer(name)
)

func HelloApiHandler(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "getHello")
	defer span.End()

	// time.Sleep(3 * time.Second)
	slog.InfoContext(c.Request().Context(), "Hi!!!")

	return c.String(http.StatusOK, "Hi!!!")
}
