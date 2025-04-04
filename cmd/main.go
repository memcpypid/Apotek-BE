package main

import (
	"apotek-management/config"
	"apotek-management/routes"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CheckDatabaseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.DB == nil {
			c.JSON(500, gin.H{"error": "Database is not connected"})
			c.Abort()
			return
		}
		c.Next()
	}
}
func setupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	router.Use(gin.Recovery())
	routes.SetupRoutes(router)
	return router
}

func main() {
	// gin.SetMode(gin.ReleaseMode)
	config.ConnectDB()
	r := setupRouter()
	r.Use(CheckDatabaseMiddleware())
	if err := r.Run("localhost:3000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
