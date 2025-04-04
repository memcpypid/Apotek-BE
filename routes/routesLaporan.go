package routes

import (
	"apotek-management/controllers"

	"github.com/gin-gonic/gin"
)

func LaporanRoutes(api *gin.RouterGroup) {
	//laporan
	api.GET("/laporan/laporan-transaksi", controllers.GetLaporanTransaksi)
	api.GET("/laporan/laporan-stok", controllers.GetLaporanStok)
	api.GET("/laporan/laporan-labarugi", controllers.GetLaporanLabaRugi)

}
