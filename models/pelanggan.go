package models

import (
	"time"
)

// Pelanggan: Data pelanggan yang berbelanja di apotek
type Pelanggan struct {
	ID        uint      `json:"id_pelanggan" gorm:"primaryKey"`
	Nama      string    `json:"nama" gorm:"type:varchar(100);not null"`
	Alamat    string    `json:"alamat" gorm:"type:text"`
	Telepon   string    `json:"telepon" gorm:"type:varchar(20)"`
	Email     string    `json:"email" gorm:"type:varchar(100)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
