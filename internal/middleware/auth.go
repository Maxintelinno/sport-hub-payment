package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 1. Validate Authorization header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid or missing Authorization header",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Parse JWT Claims (Unverified for now, please use proper secret in production)
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "failed to parse token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid token claims",
			})
		}

		// Extract userid (matching the user's provided structure)
		userID, ok := claims["userid"].(string)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "userid not found in token",
			})
		}

		// 3. Store in context
		c.Set("user_id", userID)

		return next(c)
	}
}
