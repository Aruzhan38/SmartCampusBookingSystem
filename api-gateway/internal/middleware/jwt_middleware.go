package middleware

import (
	"api-gateway/internal/client"
	"context"
	"net/http"
	"strings"

	"log"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware(userClient client.UserClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		// support both "Bearer <token>" and bare token values
		parts := strings.Fields(authHeader)
		var token string
		if len(parts) >= 2 && strings.ToLower(parts[0]) == "bearer" {
			token = parts[1]
		} else {
			token = authHeader
		}

		user, err := userClient.ValidateToken(context.Background(), token)
		if err != nil {
			// log error for debugging
			prefix := token
			if len(prefix) > 16 {
				prefix = prefix[:16] + "..."
			}
			log.Printf("JWT validation failed for token prefix='%s': %v", prefix, err)
			// return underlying error message from user service for easier debugging
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}
