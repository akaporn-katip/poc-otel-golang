package hello

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HelloApiHandler(c echo.Context) error {
	// slog.
	slog.InfoContext(c.Request().Context(), "Hi!!!")

	// c.Logger().Info("Hi!!!!")
	return c.String(http.StatusOK, "Hi!!!")
}
