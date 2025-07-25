package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/georgisomnoev/feature-flag-api/internal/auth/model"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/service"
	"github.com/labstack/echo/v4"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Service
type Service interface {
	Authenticate(context.Context, string, string) (string, error)
}

func RegisterHandlers(srv *echo.Echo, svc Service) {
	srv.POST("/auth", authenticateHandler(svc))
}

func authenticateHandler(svc Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req model.AuthRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		if req.Username == "" || req.Password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "username and password are required")
		}

		token, err := svc.Authenticate(c.Request().Context(), req.Username, req.Password)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to authenticate user")
		}

		return c.JSON(http.StatusOK, model.AuthResponse{Token: token})
	}
}
