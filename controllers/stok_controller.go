package controllers

import (
	"apotek-management/config"
	"apotek-management/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateStok(c *gin.Context) {
	var stokInput struct {
		ObatID              uint   `json:"obat_id" binding:"required"`
		StokAwal            int    `json:"stok_awal" binding:"required"`
		JumlahStokTransaksi int    `json:"jumlah_stok_transaksi" binding:"required"`
		TipeTransaksi       string `json:"tipe_transaksi" binding:"required"`
		TanggalKadaluwarsa  string `json:"tanggal_kadaluwarsa" binding:"required"`
		Lokasi              string `json:"lokasi" binding:"required"`
		Keterangan          string `json:"keterangan"`
	}

	if err := c.ShouldBindJSON(&stokInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate TipeTransaksi
	if stokInput.TipeTransaksi != "MASUK" && stokInput.TipeTransaksi != "KELUAR" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tipe_transaksi, must be 'MASUK' or 'KELUAR'"})
		return
	}

	// Parse tanggal kadaluwarsa
	tanggalKadaluwarsa, err := time.Parse("2006-01-02", stokInput.TanggalKadaluwarsa)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tanggal_kadaluwarsa format, use YYYY-MM-DD"})
		return
	}

	if tanggalKadaluwarsa.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tanggal kadaluwarsa cannot be in the past"})
		return
	}

	// Retrieve related obat
	var obat models.Obat
	if err := config.DB.First(&obat, stokInput.ObatID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Obat not found"})
		return
	}

	// Calculate stok_akhir
	stokAkhir := stokInput.StokAwal
	if stokInput.TipeTransaksi == "MASUK" {
		stokAkhir += stokInput.JumlahStokTransaksi
	} else if stokInput.TipeTransaksi == "KELUAR" {
		if stokAkhir < stokInput.JumlahStokTransaksi {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Jumlah stok transaksi exceeds stok_awal"})
			return
		}
		stokAkhir -= stokInput.JumlahStokTransaksi
	}

	stok := models.Stok{
		ObatID:              stokInput.ObatID,
		StokAwal:            0,
		StokAkhir:           stokInput.StokAwal,
		JumlahStokTransaksi: stokInput.JumlahStokTransaksi,
		TipeTransaksi:       stokInput.TipeTransaksi,
		TanggalKadaluwarsa:  tanggalKadaluwarsa,
		Lokasi:              stokInput.Lokasi,
		Keterangan:          stokInput.Keterangan,
	}

	// Save stok to database
	if err := config.DB.Create(&stok).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Stok created successfully", "data": stok})
}

func UpdateStok(c *gin.Context) {
	id := c.Param("id")

	// Cari stok lama berdasarkan ID
	var stokLama models.Stok
	if err := config.DB.First(&stokLama, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stok not found"})
		return
	}

	// Struktur untuk menerima input dari client
	var stokInput struct {
		JumlahStokTransaksi int    `json:"jumlah_stok_transaksi" binding:"required"`
		Keterangan          string `json:"keterangan"`
	}

	// Bind JSON input ke `stokInput`
	if err := c.ShouldBindJSON(&stokInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Pastikan jumlah stok transaksi lebih besar dari 0
	if stokInput.JumlahStokTransaksi <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jumlah stok transaksi harus lebih besar dari 0"})
		return
	}

	// Perhitungan stok akhir baru untuk restok
	stokAkhirBaru := stokLama.StokAkhir + stokInput.JumlahStokTransaksi

	// Buat entry stok baru untuk mencatat restok
	stokBaru := models.Stok{
		ObatID:              stokLama.ObatID,    // Gunakan ID obat dari stok lama
		StokAwal:            stokLama.StokAkhir, // Stok awal adalah stok akhir sebelumnya
		StokAkhir:           stokAkhirBaru,      // Hitung stok akhir baru
		JumlahStokTransaksi: stokInput.JumlahStokTransaksi,
		TipeTransaksi:       "MASUK",                     // Tipe transaksi diatur ke "MASUK"
		Keterangan:          stokInput.Keterangan,        // Gunakan keterangan dari input
		TanggalKadaluwarsa:  stokLama.TanggalKadaluwarsa, // Gunakan tanggal kadaluarsa stok lama
		Lokasi:              stokLama.Lokasi,
	}

	// Simpan entry stok baru ke database
	if err := config.DB.Create(&stokBaru).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new stok entry"})
		return
	}

	// Berikan respons sukses
	c.JSON(http.StatusOK, gin.H{"message": "Restok berhasil", "data": stokBaru})
}

func DeleteStok(c *gin.Context) {
	id := c.Param("id")

	// Cari stok yang ingin dihapus
	var stokToDelete models.Stok
	if err := config.DB.First(&stokToDelete, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stok not found"})
		return
	}

	// Ambil semua stok terkait obat_id, diurutkan berdasarkan waktu pembuatan (created_at)
	var allStok []models.Stok
	if err := config.DB.Where("obat_id = ?", stokToDelete.ObatID).Order("created_at asc").Find(&allStok).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch related stok: " + err.Error()})
		return
	}

	// Hapus stok yang ditargetkan
	if err := config.DB.Delete(&stokToDelete).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete stok: " + err.Error()})
		return
	}

	// Perhitungan ulang stok
	var stokAwal int
	if len(allStok) > 0 && allStok[0].ID != stokToDelete.ID {
		stokAwal = allStok[0].StokAwal // Ambil stok awal dari entry pertama
	} else {
		stokAwal = 0 // Jika tidak ada data awal atau stok pertama yang dihapus
	}

	for i, stok := range allStok {
		if stok.ID == stokToDelete.ID {
			// Lewati stok yang sudah dihapus
			continue
		}

		if i == 0 {
			stok.StokAwal = stokAwal
		} else {
			stok.StokAwal = allStok[i-1].StokAkhir
		}

		// Hitung stok akhir berdasarkan tipe transaksi
		if stok.TipeTransaksi == "MASUK" {
			stok.StokAkhir = stok.StokAwal + stok.JumlahStokTransaksi
		} else if stok.TipeTransaksi == "KELUAR" {
			stok.StokAkhir = stok.StokAwal - stok.JumlahStokTransaksi
		}

		// Update stok di database
		if err := config.DB.Save(&stok).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stok: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stok deleted and recalculated successfully"})
}
func GetAllStok(c *gin.Context) {
	var stokList []models.Stok
	if err := config.DB.Preload("Obat.Pemasok").Preload("Obat.Tags").Preload("Obat.TipeObat").Preload("Obat.Stok").Find(&stokList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stokList})
}

func GetStokByID(c *gin.Context) {
	id := c.Param("id")
	var stok models.Stok
	if err := config.DB.Preload("Obat").First(&stok, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stok not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stok})
}
