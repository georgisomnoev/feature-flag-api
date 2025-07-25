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

	if !validateScopes(claims["scopes"], requiredScope) {
		return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
	}

	c.Set("user_id", userUUID)
	c.Set("scopes", claims["scopes"])
	return nil
}

func validateScopes(scopes interface{}, requiredScope string) bool {
	scopeList, ok := scopes.([]interface{})
	if !ok {
		if scopeStrList, okStr := scopes.([]string); okStr {
			for _, scope := range scopeStrList {
				if scope == requiredScope {
					return true
				}
			}
		} else {
			return false
		}
	}

	for _, scope := range scopeList {
		if scope == requiredScope {
			return true
		}
	}

	return false
}
