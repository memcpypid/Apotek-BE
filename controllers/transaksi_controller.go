package controllers

import (
	"apotek-management/config"
	"apotek-management/models"
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateTransaksiInput struct {
	KodeTransaksi string                       `json:"kode_transaksi" binding:"required"`
	PelangganID   *uint                        `json:"id_pelanggan"`
	HandledBy     uint                         `json:"handled_by" binding:"required"`
	Diskon        uint64                       `json:"diskon"`
	MetodeBayar   string                       `json:"metode_bayar" binding:"required,oneof='CASH' 'TRANSFER' 'CREDIT_CARD'"`
	Status        string                       `json:"status" binding:"required,oneof='PENDING' 'SUCCESS' 'CANCEL'"`
	Obats         []CreateTransaksiDetailInput `json:"obats" binding:"required"`
}
type CreateTransaksiDetailInput struct {
	ObatID uint   `json:"id_obat" binding:"required"`
	Jumlah int    `json:"jumlah" binding:"required,min=1"`
	Harga  uint64 `json:"harga" binding:"required"`
}
type UpdateStatusInput struct {
	Status string `json:"status" binding:"required,oneof='PENDING' 'SUCCESS' 'CANCEL'"`
}

func CreateTransaksi(c *gin.Context) {
	var input CreateTransaksiInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi User
	var user models.User
	if err := config.DB.First(&user, input.HandledBy).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User dengan ID tersebut tidak ditemukan"})
		return
	}

	// Validasi Obat dan hitung TotalHarga tanpa mengurangi stok
	var totalHarga uint64 = 0
	for _, detail := range input.Obats {
		var obat models.Obat
		if err := config.DB.First(&obat, detail.ObatID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Obat dengan ID tersebut tidak ditemukan"})
			return
		}

		// Hitung subtotal untuk setiap detail
		subtotal := uint64(detail.Jumlah) * detail.Harga
		totalHarga += subtotal
	}

	// Kurangi diskon dari total harga
	if input.Diskon > totalHarga {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Diskon lebih besar dari total harga"})
		return
	}
	totalHarga -= input.Diskon

	// Buat transaksi tanpa merubah stok obat
	transaksi := models.Transaksi{
		KodeTransaksi: input.KodeTransaksi,
		HandledBy:     input.HandledBy,
		Diskon:        input.Diskon,
		MetodeBayar:   input.MetodeBayar,
		Status:        "PENDING", // Set status awal transaksi ke PENDING
		TotalHarga:    totalHarga,
		PelangganID:   input.PelangganID,
	}

	// Tambahkan obats ke transaksi tanpa merubah stok
	for _, detail := range input.Obats {
		transaksi.Obats = append(transaksi.Obats, models.TransaksiDetail{
			ObatID:   detail.ObatID,
			Jumlah:   detail.Jumlah,
			Harga:    detail.Harga,
			Subtotal: uint64(detail.Jumlah) * detail.Harga, // Hitung subtotal
		})
	}

	// Simpan transaksi
	if err := config.DB.Create(&transaksi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan saat menyimpan transaksi"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": transaksi})
}
func UpdateTransaksiStatus(c *gin.Context) {
	id := c.Param("id")            // ID transaksi dari parameter URL
	status := c.PostForm("status") // Get status from multipart form

	var transaksi models.Transaksi
	if err := config.DB.Preload("Obats").First(&transaksi, id).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Handle status update logic
		if status == "SUCCESS" && transaksi.Status != "SUCCESS" {
			for _, detail := range transaksi.Obats {
				var stok models.Stok
				if err := tx.Where("obat_id = ?", detail.ObatID).Order("id_stok DESC").First(&stok).Error; err != nil {
					return errors.New("stok obat tidak ditemukan")
				}

				newStok := models.Stok{
					ObatID:              detail.ObatID,
					Lokasi:              stok.Lokasi,
					TanggalKadaluwarsa:  stok.TanggalKadaluwarsa,
					StokAwal:            stok.StokAkhir,
					StokAkhir:           stok.StokAkhir - detail.Jumlah,
					JumlahStokTransaksi: detail.Jumlah,
					TipeTransaksi:       "KELUAR",
					Keterangan:          "Stok dikurangi untuk transaksi " + transaksi.KodeTransaksi,
				}

				if err := tx.Create(&newStok).Error; err != nil {
					return errors.New("gagal memperbarui stok")
				}
			}
			transaksi.Status = "SUCCESS"

		} else if status == "CANCEL" {
			if transaksi.Status == "SUCCESS" {
				for _, detail := range transaksi.Obats {
					var stok models.Stok
					if err := tx.Where("obat_id = ?", detail.ObatID).Order("id_stok DESC").First(&stok).Error; err != nil {
						return errors.New("stok obat tidak ditemukan")
					}

					// Restore stock
					newStok := models.Stok{
						ObatID:              detail.ObatID,
						Lokasi:              stok.Lokasi,
						TanggalKadaluwarsa:  stok.TanggalKadaluwarsa,
						StokAwal:            stok.StokAkhir,
						StokAkhir:           stok.StokAkhir + detail.Jumlah,
						JumlahStokTransaksi: detail.Jumlah,
						TipeTransaksi:       "MASUK",
						Keterangan:          "Stok dikembalikan untuk transaksi " + transaksi.KodeTransaksi,
					}

					if err := tx.Create(&newStok).Error; err != nil {
						return errors.New("gagal mengembalikan stok")
					}
				}
			}
			transaksi.Status = "CANCEL"
		}

		// **Ensure the upload directory exists**
		uploadDir := "uploads/bukti_pembayaran/"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
				return errors.New("gagal membuat folder untuk bukti pembayaran")
			}
		}

		// **Handle file upload**
		file, fileErr := c.FormFile("bukti_pembayaran")
		if fileErr == nil { // Only process if file is uploaded
			uploadPath := uploadDir + file.Filename
			if err := c.SaveUploadedFile(file, uploadPath); err != nil {
				return errors.New("gagal menyimpan bukti pembayaran")
			}

			// **Update the bukti_pembayaran path**
			if err := tx.Model(&transaksi).Update("bukti_pembayaran", uploadPath).Error; err != nil {
				return errors.New("gagal memperbarui bukti pembayaran")
			}
		}

		// Save transaction updates
		if err := tx.Save(&transaksi).Error; err != nil {
			return errors.New("gagal memperbarui status transaksi")
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transaksi})
}

// func UpdateTransaksiStatus(c *gin.Context) {
// 	id := c.Param("id") // ID transaksi dari parameter URL
// 	var input UpdateStatusInput
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	var transaksi models.Transaksi
// 	if err := config.DB.Preload("Obats").First(&transaksi, id).Error; err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaksi tidak ditemukan"})
// 		return
// 	}

// 	// Debugging: Periksa apakah transaksi memiliki obat
// 	fmt.Println("Jumlah Obat dalam Transaksi:", len(transaksi.Obats))

// 	err := config.DB.Transaction(func(tx *gorm.DB) error {
// 		if input.Status == "SUCCESS" {
// 			for _, detail := range transaksi.Obats {
// 				var stok models.Stok
// 				if err := tx.Where("obat_id = ?", detail.ObatID).Order("id_stok DESC").First(&stok).Error; err != nil {
// 					return errors.New("stok obat tidak ditemukan")
// 				}

// 				// Debugging: Print data stok sebelum diperbarui
// 				fmt.Printf("ObatID: %d, Stok Sebelum: %d, Jumlah Dipesan: %d\n", detail.ObatID, stok.StokAkhir, detail.Jumlah)

// 				// Buat entri stok baru
// 				newStok := models.Stok{
// 					ObatID:              detail.ObatID,
// 					Lokasi:              stok.Lokasi,
// 					TanggalKadaluwarsa:  stok.TanggalKadaluwarsa,
// 					StokAwal:            stok.StokAkhir,
// 					StokAkhir:           stok.StokAkhir - detail.Jumlah,
// 					JumlahStokTransaksi: detail.Jumlah,
// 					TipeTransaksi:       "KELUAR",
// 					Keterangan:          "Stok dikurangi untuk transaksi " + transaksi.KodeTransaksi,
// 				}

// 				if err := tx.Create(&newStok).Error; err != nil {
// 					return errors.New("gagal memperbarui stok")
// 				}

// 				// Debugging: Print data stok setelah diperbarui
// 				fmt.Printf("ObatID: %d, Stok Setelah: %d\n", detail.ObatID, newStok.StokAkhir)
// 			}

// 			transaksi.Status = "SUCCESS"
// 		} else if input.Status == "CANCEL" {
// 			if transaksi.Status == "SUCCESS" {
// 				for _, detail := range transaksi.Obats {
// 					var stok models.Stok
// 					if err := tx.Where("obat_id = ?", detail.ObatID).Order("id_stok DESC").First(&stok).Error; err != nil {
// 						return errors.New("stok obat tidak ditemukan")
// 					}

// 					// Debugging: Print data stok sebelum dikembalikan
// 					fmt.Printf("ObatID: %d, Stok Sebelum Dibatalkan: %d\n", detail.ObatID, stok.StokAkhir)

// 					newStok := models.Stok{
// 						ObatID:              detail.ObatID,
// 						Lokasi:              stok.Lokasi,
// 						TanggalKadaluwarsa:  stok.TanggalKadaluwarsa,
// 						StokAwal:            stok.StokAkhir,
// 						StokAkhir:           stok.StokAkhir + detail.Jumlah,
// 						JumlahStokTransaksi: detail.Jumlah,
// 						TipeTransaksi:       "MASUK",
// 						Keterangan:          "Stok dikembalikan untuk transaksi " + transaksi.KodeTransaksi,
// 					}

// 					if err := tx.Create(&newStok).Error; err != nil {
// 						return errors.New("gagal mengembalikan stok")
// 					}

// 					// Debugging: Print data stok setelah dikembalikan
// 					fmt.Printf("ObatID: %d, Stok Setelah Dibatalkan: %d\n", detail.ObatID, newStok.StokAkhir)
// 				}
// 			}
// 			transaksi.Status = "CANCEL"
// 		}

// 		if err := tx.Save(&transaksi).Error; err != nil {
// 			return errors.New("gagal memperbarui status transaksi")
// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"data": transaksi})
// }

func UpdateDataTransaksi(c *gin.Context) {
	id := c.Param("id") // ID transaksi dari parameter URL
	var transaksiBaru models.Transaksi
	var transaksiLama models.Transaksi

	// 1. Ambil data transaksi lama dengan preload Obats
	if err := config.DB.Preload("Obats").First(&transaksiLama, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	// 2. Bind data transaksi baru dari request
	if err := c.ShouldBindJSON(&transaksiBaru); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Perbarui detail obat dalam transaksi dan hitung total harga baru
	var totalHarga uint64 = 0
	for _, detailBaru := range transaksiBaru.Obats {
		for i, detailLama := range transaksiLama.Obats {
			if detailLama.ObatID == detailBaru.ObatID {
				// Perbarui jumlah, harga, dan subtotal
				transaksiLama.Obats[i].Jumlah = detailBaru.Jumlah
				transaksiLama.Obats[i].Harga = detailBaru.Harga
				transaksiLama.Obats[i].Subtotal = uint64(detailBaru.Jumlah) * detailBaru.Harga
				totalHarga += transaksiLama.Obats[i].Subtotal
			}
		}
	}

	// 4. Perbarui hanya data transaksi utama dengan total harga yang diperbarui
	transaksiLama.TotalHarga = totalHarga - transaksiBaru.Diskon
	if err := config.DB.Model(&transaksiLama).Updates(map[string]interface{}{
		"total_harga":      transaksiLama.TotalHarga,
		"metode_bayar":     transaksiBaru.MetodeBayar,
		"diskon":           transaksiBaru.Diskon,
		"status":           transaksiBaru.Status,
		"bukti_pembayaran": transaksiBaru.Gambar,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui transaksi: " + err.Error()})
		return
	}

	// 5. Simpan perubahan ke detail transaksi
	for _, detail := range transaksiLama.Obats {
		if err := config.DB.Model(&detail).Updates(map[string]interface{}{
			"jumlah":   detail.Jumlah,
			"harga":    detail.Harga,
			"subtotal": detail.Subtotal,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui detail transaksi: " + err.Error()})
			return
		}
	}

	// Ambil data transaksi terbaru
	var transaksiUpdated models.Transaksi
	config.DB.Preload("Obats").First(&transaksiUpdated, id)

	// Kembalikan respon sukses
	c.JSON(http.StatusOK, gin.H{"data": transaksiUpdated})
}

func GetAllTransaksi(c *gin.Context) {
	var transaksiList []models.Transaksi

	if err := config.DB.
		Preload("Obats").
		Preload("Obats.Obat.Tags").
		Preload("Obats.Obat.TipeObat").Preload("Obats.Obat.Stok").Preload("Obats.Obat.Stok").
		Preload("Pelanggan").
		Find(&transaksiList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaksiList)
}

func GetTransaksiByID(c *gin.Context) {
	id := c.Param("id")
	var transaksi models.Transaksi
	if err := config.DB.Preload("Obats.Obat.Tags").Preload("Obats.Obat.TipeObat").First(&transaksi, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi not found"})
		return
	}

	c.JSON(http.StatusOK, transaksi)
}

func DeleteTransaksi(c *gin.Context) {
	id := c.Param("id")
	var transaksi models.Transaksi

	// Cari transaksi berdasarkan ID
	if err := config.DB.Preload("Obats").First(&transaksi, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	// Gunakan transaksi database agar rollback jika terjadi error
	if err := config.DB.Session(&gorm.Session{SkipHooks: true}).Transaction(func(tx *gorm.DB) error {
		// **Hapus transaksi dan detail transaksi tanpa memicu hook**
		if err := tx.Session(&gorm.Session{SkipHooks: true}).Where("transaksi_id = ?", transaksi.ID).Delete(&models.TransaksiDetail{}).Error; err != nil {
			return errors.New("gagal menghapus detail transaksi: " + err.Error())
		}
		if err := tx.Session(&gorm.Session{SkipHooks: true}).Delete(&transaksi).Error; err != nil {
			return errors.New("gagal menghapus transaksi: " + err.Error())
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus transaksi: " + err.Error()})
		return
	}

	// Kembalikan respon sukses
	c.JSON(http.StatusOK, gin.H{"message": "Transaksi berhasil dihapus"})
}
