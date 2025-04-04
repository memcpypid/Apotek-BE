package controllers

import (
	"apotek-management/config"
	"apotek-management/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// **1. Tambahkan Pelanggan Baru**
func CreatePelanggan(c *gin.Context) {
	var pelanggan models.Pelanggan

	// Bind JSON input ke struct Pelanggan
	if err := c.ShouldBindJSON(&pelanggan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simpan ke database
	if err := config.DB.Create(&pelanggan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambahkan pelanggan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pelanggan berhasil ditambahkan", "data": pelanggan})
}

// **2. Ambil Semua Pelanggan**
// func GetAllPelanggan(c *gin.Context) {
// 	var pelanggan []models.Pelanggan
// 	var transaksi []models.Transaksi

// 	// Ambil semua data pelanggan dari database
// 	if err := config.DB.Find(&pelanggan).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pelanggan"})
// 		return
// 	}
// 	if err := config.DB.Find(&transaksi).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pelanggan"})
// 		return
// 	}

//		c.JSON(http.StatusOK, gin.H{"data": pelanggan, "transaksi": transaksi})
//	}
func GetAllPelanggan(c *gin.Context) {
	var pelangganList []models.Pelanggan
	var result []map[string]interface{}

	// Mengambil semua data pelanggan
	if err := config.DB.Find(&pelangganList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pelanggan"})
		return
	}

	// Iterasi setiap pelanggan dan mencari transaksinya
	for _, pelanggan := range pelangganList {
		var transaksi []models.Transaksi

		// Cari transaksi yang terkait dengan pelanggan
		config.DB.Where("pelanggan_id = ?", pelanggan.ID).Preload("Obats.Obat").Preload("Obats.Obat.TipeObat").Preload("Obats.Obat.Tags").Find(&transaksi)

		// Membentuk struktur data yang akan dikembalikan
		pelangganData := map[string]interface{}{
			"id_pelanggan": pelanggan.ID,
			"nama":         pelanggan.Nama,
			"alamat":       pelanggan.Alamat,
			"telepon":      pelanggan.Telepon,
			"email":        pelanggan.Email,
			"transaksi":    transaksi, // Jika tidak ada transaksi, ini akan menjadi array kosong []
		}

		result = append(result, pelangganData)
	}

	// Return data pelanggan beserta transaksi dalam bentuk JSON
	c.JSON(http.StatusOK, gin.H{"data": result})
}
func GetAllPelangganWithTransaksi(c *gin.Context) {
	var pelangganList []models.Pelanggan
	var result []map[string]interface{}

	// Mengambil semua data pelanggan
	if err := config.DB.Find(&pelangganList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pelanggan"})
		return
	}

	// Iterasi setiap pelanggan dan mencari transaksinya
	for _, pelanggan := range pelangganList {
		var transaksi []models.Transaksi

		// Cari transaksi yang terkait dengan pelanggan
		config.DB.Where("pelanggan_id = ?", pelanggan.ID).Preload("Obats").Find(&transaksi)

		// Membentuk struktur data yang akan dikembalikan
		pelangganData := map[string]interface{}{
			"id_pelanggan": pelanggan.ID,
			"nama":         pelanggan.Nama,
			"alamat":       pelanggan.Alamat,
			"telepon":      pelanggan.Telepon,
			"transaksi":    transaksi, // Jika tidak ada transaksi, ini akan menjadi array kosong []
		}

		result = append(result, pelangganData)
	}

	// Return data pelanggan beserta transaksi dalam bentuk JSON
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// **3. Ambil Pelanggan Berdasarkan ID**
func GetPelangganByID(c *gin.Context) {
	id := c.Param("id")
	var pelanggan models.Pelanggan

	// Cek apakah pelanggan ada
	if err := config.DB.First(&pelanggan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pelanggan tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pelanggan})
}

// **4. Update Data Pelanggan**
func UpdatePelanggan(c *gin.Context) {
	id := c.Param("id")
	var pelanggan models.Pelanggan

	// Cek apakah pelanggan ada
	if err := config.DB.First(&pelanggan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pelanggan tidak ditemukan"})
		return
	}

	// Bind JSON input ke struct Pelanggan
	if err := c.ShouldBindJSON(&pelanggan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update data pelanggan
	if err := config.DB.Model(&pelanggan).Where("id = ?", id).Updates(pelanggan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui pelanggan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pelanggan berhasil diperbarui", "data": pelanggan})
}

// **5. Hapus Pelanggan**
func DeletePelanggan(c *gin.Context) {
	id := c.Param("id")
	var pelanggan models.Pelanggan

	// Cek apakah pelanggan ada
	if err := config.DB.First(&pelanggan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pelanggan tidak ditemukan"})
		return
	}

	// Hapus pelanggan
	if err := config.DB.Delete(&pelanggan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus pelanggan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pelanggan berhasil dihapus"})
}
