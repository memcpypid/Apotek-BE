package routes

import (
	"apotek-management/controllers"

	"github.com/gin-gonic/gin"
)

func PelangganRoutes(api *gin.RouterGroup) {
	//laporan
	api.GET("/pelanggan", controllers.GetAllPelanggan)
	api.GET("/laporan/:id", controllers.GetPelangganByID)
	api.POST("/pelanggan", controllers.CreatePelanggan)
	api.PUT("/pelanggan/:id", controllers.UpdatePelanggan)
	api.DELETE("/pelanggan/:id", controllers.DeletePelanggan)
}
