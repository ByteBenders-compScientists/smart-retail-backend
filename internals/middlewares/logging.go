// logging middleware
package middlewares

import (
	"time"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/utils"
	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		startTime := time.Now()
		c.Set("start_time", startTime)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Log request details
		entry := utils.Logger.WithFields(map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status":      c.Writer.Status(),
			"duration":    duration.String(),
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"query":       c.Request.URL.RawQuery,
		})

		// Get user info if available from JWT middleware
		if userID, exists := c.Get("userID"); exists {
			entry = entry.WithField("user_id", userID)
		}
		if userRole, exists := c.Get("userRole"); exists {
			entry = entry.WithField("user_role", userRole)
		}

		// Log based on status code
		switch {
		case c.Writer.Status() >= 500:
			entry.Error("Internal server error")
		case c.Writer.Status() >= 400:
			entry.Warn("Client error")
		default:
			entry.Info("Request processed successfully")
		}
	}
}
