package routes

import (
	"apotek-management/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	protected := router.Group("/api")

	// protected.Use(middleware.AuthMiddleware())
	// {
	TagObatRoutes(protected)
	TipeObatRoutes(protected)
	StokRoutes(protected)
	TransaksiRoutes(protected)
	ObatRoutes(protected)
	LaporanRoutes(protected)
	PemasokRoutes(protected)
	userRoutes(protected)
	PelangganRoutes(protected)

	// }
	router.POST("/api/login", controllers.Login)
	router.POST("/api/signup", controllers.CreateUser)
	router.Static("/uploads", "./uploads")
	router.Static("/gambar-obat", "./gambar-obat")
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the Apotek Management API!",
		})
	})
}
