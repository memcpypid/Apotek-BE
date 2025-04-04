package routes

import (
	"apotek-management/controllers"

	"github.com/gin-gonic/gin"
)

func userRoutes(api *gin.RouterGroup) {
	// Transaksi
	api.POST("/user", controllers.CreateUser)       // Create User
	api.GET("/user", controllers.GetAllUsers)       // Get All Users
	api.GET("/user/:id", controllers.GetUserByID)   // Get User by ID
	api.PUT("/user/:id", controllers.UpdateUser)    // Update User
	api.DELETE("/user/:id", controllers.DeleteUser) // Delete User
}
