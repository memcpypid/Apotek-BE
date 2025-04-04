package routes

import (
	"apotek-management/controllers"

	"github.com/gin-gonic/gin"
)

func TipeObatRoutes(api *gin.RouterGroup) {
	// Tipe Obat
	api.GET("/tipe_obat", controllers.GetAllTipeObat)
	api.GET("/tipe_obat/:id", controllers.GetTipeObatByID)
	api.POST("/tipe_obat", controllers.CreateTipeObat)
	api.PUT("/tipe_obat/:id", controllers.UpdateTipeObat)
	api.DELETE("/tipe_obat/:id", controllers.DeleteTipeObat)
}
