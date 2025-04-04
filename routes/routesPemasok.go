package routes

import (
	"apotek-management/controllers"

	"github.com/gin-gonic/gin"
)

func PemasokRoutes(api *gin.RouterGroup) {
	//Pemasok
	api.POST("/pemasok", controllers.CreatePemasok)
	api.GET("/pemasok", controllers.GetAllPemasok)
	api.GET("/pemasok/:id", controllers.GetPemasokByID)
	api.PUT("/pemasok/:id", controllers.UpdatePemasok)
	api.DELETE("/pemasok/:id", controllers.DeletePemasok)
}
