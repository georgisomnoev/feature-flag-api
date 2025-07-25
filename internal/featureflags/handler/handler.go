package handler

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . AuthStore
type AuthStore interface {
	UserExists(ctx context.Context, id uuid.UUID) (bool, error)
}

//counterfeiter:generate . JWTHelper
type JWTHelper interface {
	ValidateToken(token string) (jwt.MapClaims, error)
}

//counterfeiter:generate . Service
type Service interface {
	ListFlags(c echo.Context) error
	GetFlagByID(c echo.Context) error

	CreateFlag(c echo.Context) error
	UpdateFlag(c echo.Context) error
	DeleteFlag(c echo.Context) error
}

func RegisterHandlers(
	ctx context.Context,
	srv *echo.Echo,
	authStore AuthStore,
	jwtHelper JWTHelper,
	svc Service,
) {
	authMiddleware := CreateAuthMiddleware(ctx, authStore, jwtHelper)

	editorGroup := srv.Group("/flags")
	editorGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("required_scope", "write:flags")
			return authMiddleware(next)(c)
		}
	})
	editorGroup.POST("", svc.CreateFlag)
	editorGroup.PUT("/:id", svc.UpdateFlag)
	editorGroup.DELETE("/:id", svc.DeleteFlag)

	viewerGroup := srv.Group("/flags")
	viewerGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("required_scope", "read:flags")
			return authMiddleware(next)(c)
		}
	})
	viewerGroup.GET("", svc.ListFlags)
	viewerGroup.GET("/:id", svc.GetFlagByID)
}
