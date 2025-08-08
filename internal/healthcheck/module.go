package healthcheck

import (
	"github.com/georgisomnoev/feature-flag-api/internal/healthcheck/handler"
	"github.com/georgisomnoev/feature-flag-api/internal/healthcheck/service"
	"github.com/labstack/echo/v4"
)

func Process(e *echo.Echo, components ...service.Component) {
	hcService := service.NewService(components)
	handler.RegisterHandlers(e, hcService)
}
