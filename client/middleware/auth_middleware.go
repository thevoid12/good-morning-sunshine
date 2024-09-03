package middleware

import (
	"context"
	"gms/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(ctx context.Context) gin.HandlerFunc {

	return func(c *gin.Context) {
		// Attach the context to the request
		tokenString := c.Query("tkn")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "JWT is required"})
			return
		}
		_, err := auth.VerifyJWTToken(ctx, tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid JWT token"})
			return
		}
		c.Next()
	}
}
