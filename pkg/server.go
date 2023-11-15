package pkg

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Run(addr string) {
	// Initialize Gin router
	r := gin.Default()
	// Set up file upload route
	r.POST("/upload", Upload)
	r.GET("/download", Download)

	// Start the Gin server
	log.Println("Server is running on :8080...")
	r.Run(addr)

}
