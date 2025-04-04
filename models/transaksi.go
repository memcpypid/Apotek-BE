package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Transaksi struct {
	ID            uint              `json:"id_transaksi" gorm:"primaryKey"`
	KodeTransaksi string            `json:"kode_transaksi" gorm:"type:varchar(20);unique;not null"`
	PelangganID   *uint             `json:"id_pelanggan" gorm:"null"`
	Pelanggan     *Pelanggan        `json:"pelanggan" gorm:"foreignKey:PelangganID;references:ID"`
	HandledBy     uint              `json:"handled_by" gorm:"not null"`
	User          User              `json:"user" gorm:"foreignKey:HandledBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	TotalHarga    uint64            `json:"total_harga" gorm:"not null"`
	Diskon        uint64            `json:"diskon" gorm:"default:0"`
	MetodeBayar   string            `json:"metode_bayar" gorm:"type:enum('CASH', 'TRANSFER', 'DEBIT','QRIS');not null"`
	Status        string            `json:"status" gorm:"type:enum('PENDING', 'SUCCESS', 'CANCEL');not null"`
	Gambar        string            `json:"bukti_pembayaran" gorm:"column:bukti_pembayaran;type:varchar(255)"`
	Obats         []TransaksiDetail `json:"obats" gorm:"foreignKey:TransaksiID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt     time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

func (t *Transaksi) CalculateTotal(tx *gorm.DB) error {
	var total uint64 = 0

	// Ambil semua TransaksiDetail terkait transaksi ini
	var details []TransaksiDetail
	if err := tx.Where("transaksi_id = ?", t.ID).Find(&details).Error; err != nil {
		return err
	}

	// Hitung total harga dari semua detail
	for _, detail := range details {
		if detail.Subtotal > (uint64(^uint(0)) - total) { // Deteksi overflow
			return errors.New("total harga mengalami overflow")
		}
		total += detail.Subtotal
	}

	// Kurangi diskon
	if t.Diskon > total {
		return errors.New("diskon lebih besar dari total harga")
	}
	t.TotalHarga = total - t.Diskon

	return nil
}

// Hook `BeforeCreate`: Validasi sebelum transaksi dibuat
func (t *Transaksi) BeforeCreate(tx *gorm.DB) (err error) {
	// Validasi jika tidak ada detail obat
	if len(t.Obats) == 0 {
		return errors.New("transaksi harus memiliki minimal 1 obat")
	}
	return nil
}

// Hook `BeforeUpdate`: Validasi atau perhitungan sebelum transaksi diperbarui
func (t *Transaksi) BeforeUpdate(tx *gorm.DB) (err error) {
	return t.CalculateTotal(tx)
}
