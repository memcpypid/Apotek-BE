package models

import "time"

// Pemasok: Informasi pemasok obat
type Pemasok struct {
	ID        uint      `json:"id_pemasok" gorm:"primaryKey"`
	Nama      string    `json:"nama" gorm:"type:varchar(100);not null"`
	Alamat    string    `json:"alamat" gorm:"type:text"`
	Telepon   string    `json:"telepon" gorm:"type:varchar(20);not null"`
	Email     string    `json:"email" gorm:"type:varchar(100)"`
	Obats     []Obat    `json:"obats" gorm:"foreignKey:PemasokID"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
