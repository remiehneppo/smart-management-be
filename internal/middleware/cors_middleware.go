package middleware

import "github.com/gin-gonic/gin"

func CorsMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With, Accept")
	c.Writer.Header().Set("Content-Type", "application/json")
	// c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}
	c.Next()
}
