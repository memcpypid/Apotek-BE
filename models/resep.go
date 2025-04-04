package models

import (
	"time"
)

type Resep struct {
	ID           uint      `json:"id_resep" gorm:"primaryKey"`
	KodeResep    string    `json:"kode_resep" gorm:"type:varchar(50);unique;not null"`
	NamaDokter   string    `json:"nama_dokter" gorm:"type:varchar(100);not null"`
	PelangganID  uint      `json:"id_pelanggan" gorm:"not null"`
	Pelanggan    Pelanggan `json:"pelanggan" gorm:"foreignKey:PelangganID;references:ID"`
	KandungaObat string    `json:"kandungan_obat" gorm:"type:text"`
	Deskripsi    string    `json:"deskripsi" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Resep) TableName() string {
	return "resep"
}
