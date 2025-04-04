package routes

import (
	"apotek-management/controllers"

	"github.com/gin-gonic/gin"
)

func TransaksiRoutes(api *gin.RouterGroup) {
	// Transaksi
	api.POST("/transaksi", controllers.CreateTransaksi)
	api.GET("/transaksi", controllers.GetAllTransaksi)
	api.GET("/transaksi/:id", controllers.GetTransaksiByID)
	api.PUT("/transaksi/:id", controllers.UpdateDataTransaksi)
	api.PUT("/transaksistatus/:id", controllers.UpdateTransaksiStatus)
	api.DELETE("/transaksi/:id", controllers.DeleteTransaksi)
}
