package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/georgisomnoev/feature-flag-api/internal/auth/model"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/service"
	"github.com/labstack/echo/v4"
)

//counterfeiter:generate . Service
type Service interface {
	Authenticate(context.Context, string, string) (string, error)
}

func RegisterHandlers(ctx context.Context, srv *echo.Echo, svc Service) {
	srv.POST("/auth", handleAuthenticion(ctx, svc))
}

func handleAuthenticion(ctx context.Context, svc Service) func(c echo.Context) error {
	return func(c echo.Context) error {
		var req model.AuthRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "invalid request",
			})
		}

		token, err := svc.Authenticate(ctx, req.Username, req.Password)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "invalid credentials",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "an error occurred while processing your request",
				"error":   err.Error(),
			})
		}

		return c.JSON(http.StatusOK, model.AuthResponse{Token: token})
	}
}
