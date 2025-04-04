package routes

import (
	"apotek-management/controllers"

	"github.com/gin-gonic/gin"
)

func TagObatRoutes(api *gin.RouterGroup) {
	// Tag Obat
	api.GET("/tag_obat", controllers.GetAllTagObat)
	api.GET("/tag_obat/:id", controllers.GetTagObatByID)
	api.POST("/tag_obat", controllers.CreateTagObat)
	api.PUT("/tag_obat/:id", controllers.UpdateTagObat)
	api.DELETE("/tag_obat/:id", controllers.DeleteTagObat)
}
