package routes

import (
	"apotek-management/controllers"

	"github.com/gin-gonic/gin"
)

func ObatRoutes(api *gin.RouterGroup) {
	// Obat
	api.POST("/obat", controllers.CreateObat)
	api.POST("/obat/batch", controllers.CreateBatchObat)
	api.GET("/obat/cari-gambar/:nama_obat", controllers.FetchDataFromGraphQL)
	api.GET("/obat", controllers.GetAllObat)
	api.GET("/obat/:id", controllers.GetObatByID)

	api.PUT("/obat/:id", controllers.UpdateObat)

	api.DELETE("/obat/:id", controllers.DeleteObat)
}
