package models

import "time"

type Stok struct {
	ID                  uint      `json:"id_stok" gorm:"primaryKey;column:id_stok"`
	ObatID              uint      `json:"obat_id" gorm:"not null"`
	Obat                Obat      `json:"obat" gorm:"foreignKey:ObatID;references:ID"`
	Lokasi              string    `json:"lokasi" gorm:"type:varchar(100);not null"` // Lokasi stok (gudang/cabang)
	TanggalKadaluwarsa  time.Time `json:"tanggal_kadaluwarsa" gorm:"not null"`      // Expiration date
	StokAwal            int       `json:"stok_awal" gorm:"not null"`
	StokAkhir           int       `json:"stok_akhir" gorm:"not null"`
	JumlahStokTransaksi int       `json:"jumlah_stok_transaksi" gorm:"not null"`
	TipeTransaksi       string    `json:"tipe_transaksi" gorm:"type:enum('MASUK', 'KELUAR');not null"`
	Keterangan          string    `json:"keterangan" gorm:"type:text"`
	CreatedAt           time.Time `json:"created_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}
