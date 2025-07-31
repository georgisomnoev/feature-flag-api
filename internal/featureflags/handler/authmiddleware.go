package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func createAuthMiddleware(authStore AuthStore, jwtHelper JWTHelper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := extractToken(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			claims, err := jwtHelper.ValidateToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}
			if err := validateTokenClaims(c.Request().Context(), c, claims, authStore); err != nil {
				return err
			}

			return next(c)
		}
	}
}

func extractToken(c echo.Context) (string, error) {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("missing token")
	}

	return strings.TrimPrefix(token, "Bearer "), nil
}

func validateTokenClaims(ctx context.Context, c echo.Context, claims jwt.MapClaims, authStore AuthStore) error {
	userID, err := claims.GetSubject()
	if err != nil || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID in token")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil || userUUID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID in token")
	}

	exists, err := authStore.UserExists(ctx, userUUID)
	if err != nil || !exists {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	requiredScope, ok := c.Get("required_scope").(string)
	if !ok || requiredScope == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "required scope not set")
	}
	if claims["scopes"] == nil {
		return echo.NewHTTPError(http.StatusForbidden, "no scopes found in token")
	}

	normalizeScopes, err := normalizeScopes(claims["scopes"])
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid scope format")
	}
	if !validateScopes(normalizeScopes, requiredScope) {
		return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
	}

	c.Set("user_id", userUUID)
	c.Set("scopes", claims["scopes"])
	return nil
}

func normalizeScopes(scopes any) ([]string, error) {
	switch v := scopes.(type) {
	case []string:
		return v, nil
	case []any:
		strScopes := make([]string, len(v))
		for i, scope := range v {
			str, ok := scope.(string)
			if !ok {
				return nil, fmt.Errorf("invalid type in scope: %v", scope)
			}
			strScopes[i] = str
		}
		return strScopes, nil
	default:
		return nil, fmt.Errorf("unsupported scopes type: %T", scopes)
	}
}

func validateScopes(scopes []string, requiredScope string) bool {
	for _, scope := range scopes {
		if scope == requiredScope {
			return true
		}
	}

	return false
}
