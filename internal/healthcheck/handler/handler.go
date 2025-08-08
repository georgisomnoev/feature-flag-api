package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Service interface {
	Status() (map[string]string, bool)
}

func RegisterHandlers(
	srv *echo.Echo,
	hs Service,
) {
	srv.GET("/healthz", handleHealthCheck(hs))
}

func handleHealthCheck(h Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		status, ok := h.Status()
		code := http.StatusOK
		if !ok {
			code = http.StatusServiceUnavailable
		}
		return c.JSON(code, status)
	}
}
