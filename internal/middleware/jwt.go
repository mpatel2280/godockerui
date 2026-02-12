package middleware

import "github.com/gin-gonic/gin"

// JWT returns a pass-through middleware for demo mode.
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
