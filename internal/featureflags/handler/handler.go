package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//go:generate gowrap gen -g -p ./ -i AuthStore -t ../../observability/templates/otel_trace.tmpl -o ./wrapped/trace/auth_store.go
//counterfeiter:generate . AuthStore
type AuthStore interface {
	UserExists(ctx context.Context, id uuid.UUID) (bool, error)
}

//go:generate gowrap gen -g -p ./ -i JWTHelper -t ../../observability/templates/otel_trace.tmpl -o ./wrapped/trace/jwt_helper.go
//counterfeiter:generate . JWTHelper
type JWTHelper interface {
	ValidateToken(token string) (jwt.MapClaims, error)
}

//go:generate gowrap gen -g -p ./ -i Service -t ../../observability/templates/otel_trace.tmpl -o ./wrapped/trace/service.go
//counterfeiter:generate . Service
type Service interface {
	ListFlags(context.Context) ([]model.FeatureFlag, error)
	GetFlagByID(context.Context, uuid.UUID) (model.FeatureFlag, error)

	CreateFlag(context.Context, model.FeatureFlagRequest) (uuid.UUID, error)
	UpdateFlag(context.Context, uuid.UUID, model.FeatureFlagRequest) error
	DeleteFlag(context.Context, uuid.UUID) error
}

type Handler struct {
	svc       Service
	authStore AuthStore
	jwtHelper JWTHelper
}

func NewHandler(svc Service, authStore AuthStore, jwtHelper JWTHelper) *Handler {
	return &Handler{
		svc:       svc,
		authStore: authStore,
		jwtHelper: jwtHelper,
	}
}

func (h *Handler) RegisterHandlers(srv *echo.Echo) {
	authMiddleware := createAuthMiddleware(h.authStore, h.jwtHelper)

	editorGroup := srv.Group("/flags")
	editorGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("required_scope", "write:flags")
			return authMiddleware(next)(c)
		}
	})
	editorGroup.POST("", h.createFlag)
	editorGroup.PUT("/:id", h.updateFlag)
	editorGroup.DELETE("/:id", h.deleteFlag)

	viewerGroup := srv.Group("/flags")
	viewerGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("required_scope", "read:flags")
			return authMiddleware(next)(c)
		}
	})
	viewerGroup.GET("", h.listFlags)
	viewerGroup.GET("/:id", h.getFlagByID)
}

func (h *Handler) listFlags(c echo.Context) error {
	flags, err := h.svc.ListFlags(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, flags)
}

func (h *Handler) getFlagByID(c echo.Context) error {
	flagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid flag ID")
	}

	flag, err := h.svc.GetFlagByID(c.Request().Context(), flagID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "feature flag not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, flag)
}

func (h *Handler) createFlag(c echo.Context) error {
	var req model.FeatureFlagRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("missing required request fields: %w", err))
	}

	flagID, err := h.svc.CreateFlag(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id": flagID,
	})
}

func (h *Handler) updateFlag(c echo.Context) error {
	flagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid flag ID")
	}

	var req model.FeatureFlagRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("missing required request fields: %w", err))
	}

	if err := h.svc.UpdateFlag(c.Request().Context(), flagID, req); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "feature flag not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) deleteFlag(c echo.Context) error {
	flagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid flag ID")
	}

	if err := h.svc.DeleteFlag(c.Request().Context(), flagID); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "feature flag not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
