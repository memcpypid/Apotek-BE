package routes

import (
	"apotek-management/controllers"

	"github.com/gin-gonic/gin"
)

func StokRoutes(api *gin.RouterGroup) {
	// Stok
	api.GET("/stok", controllers.GetAllStok)
	api.GET("stok/:id", controllers.GetStokByID)
	api.POST("/stok", controllers.CreateStok)
	api.PUT("/stok/update/:id", controllers.UpdateStok)
	api.DELETE("/stok/:id", controllers.DeleteStok)
}
