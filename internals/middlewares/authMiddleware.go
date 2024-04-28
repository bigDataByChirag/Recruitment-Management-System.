package middlewares

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"rms/internals/auth"
	"strings"
)

// Mid is a middleware struct containing an authentication instance.
type Mid struct {
	a *auth.Auth
}

// NewMid creates a new middleware instance with the provided authentication service.
// It returns an error if the authentication service is nil.
func NewMid(a *auth.Auth) (*Mid, error) {
	if a == nil {
		return nil, errors.New("auth struct cannot be nil")
	}
	return &Mid{a: a}, nil
}

// JWTMiddleware is a middleware function that checks for a JWT token in the request header,
// verifies the token, and allows the request to proceed if the token is valid and the required role is satisfied.
// JWTMiddleware is a Gin middleware function that verifies JWT tokens
func (m Mid) JWTMiddleware(handler gin.HandlerFunc, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the JWT token from the request header
		tokenString := extractTokenFromHeader(c.Request)
		if tokenString == "" {
			// If no token is found, respond with Unauthorized status
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Verify the JWT token using the authentication service
		claims, err := m.a.VerifyToken(tokenString, requiredRole)
		if err != nil {
			// If token verification fails, respond with Unauthorized status
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Add the claims to the context for further use
		c.Set("claims", claims)

		// Call the provided handler function
		handler(c)
	}
}

// extractTokenFromHeader extracts the JWT token from the Authorization header in the request.
func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	// Split the Authorization header into parts
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	// Return the token part
	return parts[1]
}
