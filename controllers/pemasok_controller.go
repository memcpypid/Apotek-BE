package controllers

import (
	"apotek-management/config"
	"apotek-management/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreatePemasok(c *gin.Context) {
	var pemasok models.Pemasok
	if err := c.ShouldBindJSON(&pemasok); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if err := config.DB.Create(&pemasok).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pemasok: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pemasok created successfully", "data": pemasok})
}

func GetAllPemasok(c *gin.Context) {
	var pemasoks []models.Pemasok
	if err := config.DB.Preload("Obats").Preload("Obats.Stok").Preload("Obats.Tags").Preload("Obats.TipeObat").Find(&pemasoks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pemasoks: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pemasoks})
}

func GetPemasokByID(c *gin.Context) {
	id := c.Param("id")
	var pemasok models.Pemasok

	if err := config.DB.First(&pemasok, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pemasok not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pemasok})
}

func UpdatePemasok(c *gin.Context) {
	id := c.Param("id")
	var pemasok models.Pemasok

	if err := config.DB.First(&pemasok, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pemasok not found"})
		return
	}

	if err := c.ShouldBindJSON(&pemasok); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if err := config.DB.Save(&pemasok).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pemasok: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pemasok updated successfully", "data": pemasok})
}

func DeletePemasok(c *gin.Context) {
	id := c.Param("id")
	var pemasok models.Pemasok

	if err := config.DB.First(&pemasok, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pemasok not found"})
		return
	}

	if err := config.DB.Delete(&pemasok).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pemasok: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pemasok deleted successfully"})
}
