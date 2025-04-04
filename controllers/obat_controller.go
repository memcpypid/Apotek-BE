package controllers

import (
	"apotek-management/config"
	"apotek-management/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
)

func CreateObat(c *gin.Context) {
	// Retrieve input fields
	namaObat := c.PostForm("nama_obat")
	dosisObat := c.PostForm("dosis_obat")
	deskripsi := c.PostForm("deskripsi")
	kodeObat := c.PostForm("kode_obat")
	golonganObat := c.PostForm("golongan_obat")
	merkObat := c.PostForm("merk_obat")
	idTipeObat, err := strconv.Atoi(c.PostForm("id_tipe_obat"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id_tipe_obat"})
		return
	}
	hargaBeli, err := strconv.ParseUint(c.PostForm("harga_beli"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid harga_beli"})
		return
	}
	hargaJual, err := strconv.ParseUint(c.PostForm("harga_jual"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid harga_jual"})
		return
	}

	pemasokID, err := strconv.Atoi(c.PostForm("id_pemasok"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id_pemasok"})
		return
	}

	// Calculate margin
	margin := hargaJual - hargaBeli

	tagIDs := c.PostFormArray("tags[]")
	var tags []models.TagObat
	for _, tagID := range tagIDs {
		var tag models.TagObat
		if err := config.DB.First(&tag, tagID).Error; err == nil {
			tags = append(tags, tag)
		}
	}

	file, err := c.FormFile("gambar")

	var filePath string

	if err == nil {
		// ✅ CASE 1: Gambar berupa file upload
		filePath = fmt.Sprintf("uploads/obat/%d-%s", time.Now().UnixNano(), file.Filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}
	} else {
		// ✅ CASE 2: Gambar dikirim sebagai path string (misal dari mobile/web form)
		pathString := c.PostForm("gambar")

		if pathString == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Gambar harus diupload atau disertakan path-nya"})
			return
		}

		// Opsional: validasi path string, misalnya harus mengandung folder tertentu
		if !strings.HasPrefix(pathString, "gambar-obat/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Path gambar tidak valid"})
			return
		}

		filePath = pathString // langsung gunakan path string
	}

	obat := models.Obat{
		NamaObat:   namaObat,
		Dosis:      dosisObat,
		Deskripsi:  deskripsi,
		TipeObatID: uint(idTipeObat),
		PemasokID:  uint(pemasokID),
		HargaBeli:  hargaBeli,
		HargaJual:  hargaJual,
		Golongan:   golonganObat,
		Merk:       merkObat,
		Margin:     margin,
		Gambar:     filePath,
		Tags:       tags,
		KodeObat:   kodeObat,
	}

	if err := config.DB.Create(&obat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"obat": obat})
}

func GetAllObat(c *gin.Context) {
	var obatList []models.Obat
	if err := config.DB.
		Preload("TipeObat").
		Preload("Pemasok").
		Preload("Tags").
		Preload("Stok").
		Find(&obatList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, obatList)
}

func GetObatByID(c *gin.Context) {
	id := c.Param("id")
	var obat models.Obat

	if err := config.DB.
		Preload("TipeObat").
		Preload("Pemasok").
		Preload("Tags").
		Preload("Stok").
		First(&obat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Obat not found"})
		return
	}

	c.JSON(http.StatusOK, obat)
}

func UpdateObat(c *gin.Context) {
	id := c.Param("id")
	var existingObat models.Obat

	if err := config.DB.First(&existingObat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Obat not found"})
		return
	}

	var updatedObat struct {
		KodeObat    string           `json:"kode_obat"`
		NamaObat    string           `json:"nama_obat"`
		Dosis       string           `json:"dosis_obat"`
		Deskripsi   string           `json:"deskripsi"`
		HargaBeli   uint64           `json:"harga_beli"`
		HargaJual   uint64           `json:"harga_jual"`
		TipeObatID  uint             `json:"id_tipe_obat"`
		PemasokID   uint             `json:"id_pemasok"`
		ResepDokter bool             `json:"is_prescription"`
		Tags        []models.TagObat `json:"tags"`
	}

	if err := c.ShouldBindJSON(&updatedObat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Update image if new file is provided
	file, err := c.FormFile("gambar")
	if err == nil {
		if existingObat.Gambar != "" {
			_ = os.Remove(existingObat.Gambar)
		}
		filePath := fmt.Sprintf("uploads/obat/%d-%s", time.Now().UnixNano(), file.Filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image: " + err.Error()})
			return
		}
		existingObat.Gambar = filePath
	}

	existingObat.KodeObat = updatedObat.KodeObat
	existingObat.NamaObat = updatedObat.NamaObat
	existingObat.Dosis = updatedObat.Dosis
	existingObat.Deskripsi = updatedObat.Deskripsi
	existingObat.HargaBeli = updatedObat.HargaBeli
	existingObat.HargaJual = updatedObat.HargaJual
	existingObat.Margin = updatedObat.HargaJual - updatedObat.HargaBeli
	existingObat.TipeObatID = updatedObat.TipeObatID
	existingObat.PemasokID = updatedObat.PemasokID
	existingObat.IsPrescription = updatedObat.ResepDokter

	if err := config.DB.Save(&existingObat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Obat: " + err.Error()})
		return
	}

	if len(updatedObat.Tags) > 0 {
		var tagIDs []uint
		for _, tag := range updatedObat.Tags {
			tagIDs = append(tagIDs, tag.ID)
		}

		var tags []models.TagObat
		if err := config.DB.Where("id_tag_obat IN ?", tagIDs).Find(&tags).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tags: " + err.Error()})
			return
		}
		if err := config.DB.Model(&existingObat).Association("Tags").Replace(&tags); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tags association: " + err.Error()})
			return
		}
	}

	if err := config.DB.Preload("TipeObat").Preload("Pemasok").Preload("Tags").First(&existingObat, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated Obat: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, existingObat)
}

func DeleteObat(c *gin.Context) {
	id := c.Param("id")
	var obat models.Obat

	if err := config.DB.First(&obat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Obat not found"})
		return
	}

	if err := config.DB.Model(&obat).Association("Tags").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete related tags: " + err.Error()})
		return
	}

	if err := config.DB.Where("obat_id = ?", obat.ID).Delete(&models.Stok{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete related stock: " + err.Error()})
		return
	}

	if err := config.DB.Delete(&obat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete obat: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Obat and related data deleted successfully"})
}

// CreateBatchObat handles batch insertion of Obat along with image uploads
func CreateBatchObat(c *gin.Context) {
	// Parse multipart form data
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid multipart form data"})
		return
	}

	// Retrieve JSON data and image files
	files := form.File["gambar"]
	data := form.Value["data"]

	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No data received"})
		return
	}

	var obatList []models.Obat
	var obatInputList []struct {
		KodeObat  string `json:"kode_obat"`
		NamaObat  string `json:"nama_obat"`
		DosisObat string `json:"dosis_obat"`
		Deskripsi string `json:"deskripsi"`
		HargaBeli uint64 `json:"harga_beli"`
		HargaJual uint64 `json:"harga_jual"`
		Merk      string `json:"merk_obat"`
		Golongan  string `json:"golongan_obat"`
		Pemasok   struct {
			Nama string `json:"nama"`
		} `json:"pemasok"`
		TipeObat struct {
			NamaTipe string `json:"nama_tipe"`
		} `json:"tipe_obat"`
		TagObat []string `json:"tag_obat"`
		Gambar  string   `json:"gambar"`
		Resep   bool     `json:"resep"`
		Margin  uint64   `json:"margin"`
		Stok    []struct {
			Lokasi              string `json:"lokasi"`
			TanggalKadaluwarsa  string `json:"tanggal_kadaluwarsa"`
			StokAwal            int    `json:"stok_awal"`
			StokAkhir           int    `json:"stok_akhir"`
			JumlahStokTransaksi int    `json:"jumlah_stok_transaksi"`
			TipeTransaksi       string `json:"tipe_transaksi"`
			Keterangan          string `json:"keterangan"`
		} `json:"stok"`
	}

	// Parse JSON input
	if err := json.Unmarshal([]byte(data[0]), &obatInputList); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid JSON format: %v", err)})
		return
	}

	// Start a database transaction
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		// Process each obat entry
		for i, item := range obatInputList {
			// Save image file if provided
			var filePath string
			if i < len(files) {
				file := files[i]
				fileExt := filepath.Ext(file.Filename)
				fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), fileExt)
				filePath = "uploads/obat/" + fileName

				if err := c.SaveUploadedFile(file, filePath); err != nil {
					return fmt.Errorf("failed to save image: %v", err)
				}
			}

			// Find or create TipeObat
			var tipeObat models.TipeObat
			if err := tx.Where("nama_tipe = ?", item.TipeObat.NamaTipe).FirstOrCreate(&tipeObat, models.TipeObat{NamaTipe: item.TipeObat.NamaTipe}).Error; err != nil {
				return fmt.Errorf("failed to find or create tipe obat: %v", err)
			}

			// Find or create Pemasok
			var pemasok models.Pemasok
			if err := tx.Where("nama = ?", item.Pemasok.Nama).FirstOrCreate(&pemasok, models.Pemasok{Nama: item.Pemasok.Nama}).Error; err != nil {
				return fmt.Errorf("failed to find or create pemasok: %v", err)
			}

			// Find or create tags
			var tagList []models.TagObat
			for _, tagName := range item.TagObat {
				var tag models.TagObat
				if err := tx.Where("nama_tag = ?", tagName).FirstOrCreate(&tag, models.TagObat{NamaTag: tagName}).Error; err != nil {
					return fmt.Errorf("failed to find or create tag: %v", err)
				}
				tagList = append(tagList, tag)
			}

			// Create Obat instance
			obat := models.Obat{
				KodeObat:       item.KodeObat,
				NamaObat:       item.NamaObat,
				Dosis:          item.DosisObat,
				Deskripsi:      item.Deskripsi,
				HargaBeli:      item.HargaBeli,
				HargaJual:      item.HargaJual,
				Margin:         item.Margin,
				Golongan:       item.Golongan,
				Merk:           item.Merk,
				IsPrescription: item.Resep,
				PemasokID:      pemasok.ID,
				TipeObatID:     tipeObat.ID,
				Tags:           tagList,
				Gambar:         filePath,
			}

			// Save Obat to database
			if err := tx.Create(&obat).Error; err != nil {
				return fmt.Errorf("failed to save obat data: %v", err)
			}

			// Save Stok Data
			for _, stokItem := range item.Stok {
				stok := models.Stok{
					ObatID:              obat.ID,
					Lokasi:              stokItem.Lokasi,
					TanggalKadaluwarsa:  parseDate(stokItem.TanggalKadaluwarsa),
					StokAwal:            0,
					StokAkhir:           stokItem.StokAkhir,
					JumlahStokTransaksi: stokItem.JumlahStokTransaksi,
					TipeTransaksi:       stokItem.TipeTransaksi,
					Keterangan:          stokItem.Keterangan,
				}

				if err := tx.Create(&stok).Error; err != nil {
					return fmt.Errorf("failed to save stok data: %v", err)
				}
			}

			obatList = append(obatList, obat)
		}

		// Commit transaction
		return nil
	})

	// Handle transaction result
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "Batch obat created successfully",
		"data":    obatList,
	})
}

// Helper function to parse date from string (format: dd-mm-yyyy)
func parseDate(dateString string) time.Time {
	layout := "02-01-2006"
	parsedTime, err := time.Parse(layout, dateString)
	if err != nil {
		return time.Now() // Return current date if parsing fails
	}
	return parsedTime
}

type SearchProductInput struct {
	Sort       []SortField `json:"sort"`
	Filter     Filter      `json:"filter"`
	Pagination Pagination  `json:"pagination"`
}

type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

type Filter struct {
	CategoryIds       []string `json:"categoryIds"`
	Strategy          string   `json:"strategy"`
	Keyword           string   `json:"keyword"`
	TkdnBmp           *string  `json:"tkdnBmp"`
	Labels            []string `json:"labels"`
	SellerTypes       []string `json:"sellerTypes"`
	SellerRegionCodes []string `json:"sellerRegionCodes"`
	MinPrice          *int     `json:"minPrice"`
	MaxPrice          *int     `json:"maxPrice"`
	RateTypes         []string `json:"rateTypes"`
	ProductTypes      []string `json:"productTypes"`
	RatingAvgGte      *float64 `json:"ratingAvgGte"`
}

type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"perPage"`
}

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}
type GraphQLResponse struct {
	Data struct {
		SearchProducts struct {
			Items []struct {
				Images []string `json:"images"`
			} `json:"items"`
		} `json:"searchProducts"`
	} `json:"data"`
}

func FetchDataFromGraphQL(c *gin.Context) {
	nama_obat := c.Param("nama_obat")
	payload := GraphQLRequest{
		Query: "query searchProducts($input: SearchProductInput!) {\n  searchProducts(input: $input) {\n    ... on ListSearchProductResponse {\n      total\n      perPage\n      currentPage\n      lastPage\n      items {\n        id\n        type\n        isActive\n        images\n        isPreOrder\n        isRegionPrice\n        isSellerUMKK\n        labels\n        isWholesale\n        defaultPrice\n        defaultPriceWithTax\n        createdAt\n        maxPrice\n        maxPriceWithTax\n        minPrice\n        minPriceWithTax\n        ppnBmPercentage\n        ppnPercentage\n        tkdn {\n          value\n          bmpValue\n          tkdnBmp\n          status\n        }\n        location {\n          name\n          regionCode\n          child {\n            name\n            regionCode\n            child {\n              name\n              regionCode\n              child {\n                name\n                regionCode\n              }\n            }\n          }\n        }\n        name\n        stockAvailability\n        stockAccumulation\n        sellerName\n        sellerId\n        score\n        scoreDetail {\n          keywordScore\n          locationScore\n          priceScore\n          ratingScore\n          tkdnScore\n          umkkScore\n          unitSoldScore\n        }\n        unitSold\n        username\n        slug\n        rating {\n          count\n          average\n        }\n        variants {\n          id\n          isActive\n          options {\n            name\n            value\n          }\n          price\n          priceWithTax\n          sortOrder\n          stock\n        }\n        status\n      }\n    }\n    ... on GenericError {\n      __typename\n      reqId\n      message\n      code\n    }\n  }\n}",
		Variables: map[string]interface{}{
			"input": SearchProductInput{
				Sort: []SortField{
					{
						Field: "RELEVANCE",
						Order: "DESC",
					},
				},
				Filter: Filter{
					// CategoryIds:       []string{"ae9a916b-ff14-4681-9a6b-1cc853d4aab7", "40620237-5679-48a4-a990-81cd821f9f87"},
					Strategy:          "SRP",
					Keyword:           nama_obat,
					TkdnBmp:           nil,
					Labels:            []string{},
					SellerTypes:       []string{},
					SellerRegionCodes: []string{""},
					MinPrice:          nil,
					MaxPrice:          nil,
					RateTypes:         []string{},
					ProductTypes:      []string{},
					RatingAvgGte:      nil,
				},
				Pagination: Pagination{
					Page:    1,
					PerPage: 100,
				},
			},
		},
		OperationName: "searchProducts",
	}

	client := resty.New()

	// Make the GraphQL request
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post("https://katalog.inaproc.id/graphql")

	if err != nil {
		log.Printf("Error fetching data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	var result GraphQLResponse
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		log.Printf("Error parsing response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}
	// Extract images from the response
	var images []string
	for _, item := range result.Data.SearchProducts.Items {
		images = append(images, item.Images...)
	}

	// Tentukan folder penyimpanan
	dir := filepath.Join("gambar-obat", nama_obat)
	os.MkdirAll(dir, os.ModePerm)

	// Download gambar
	var savedFiles []string
	for i, item := range result.Data.SearchProducts.Items {
		for j, imgURL := range item.Images {
			// Format nama file yang benar
			fileName := fmt.Sprintf("%s_%d_%d.jpg", nama_obat, i, j)
			filePath := filepath.Join(dir, fileName)

			err := downloadFile(imgURL, filePath)
			if err != nil {
				log.Printf("Failed to download image: %v", err)
			} else {
				savedFiles = append(savedFiles, filePath)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"gambar_obat": savedFiles})
}
func downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
