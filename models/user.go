package models

import "time"

type User struct {
	ID        uint        `json:"id_user" gorm:"primaryKey;autoIncrement"`
	Username  string      `json:"username" gorm:"unique;not null"`
	Password  string      `json:"password" gorm:"not null"`
	FullName  string      `json:"full_name" gorm:"not null"`
	Role      string      `json:"role" gorm:"type:enum('kasir','apoteker');default:'apoteker';not null"`
	Email     string      `json:"email" gorm:"unique;not null"`
	Telpon    string      `json:"telpon" gorm:"type:varchar(15);unique;not null"`
	Alamat    string      `json:"alamat" gorm:"type:text"`
	Transaksi []Transaksi `json:"transaksi" gorm:"foreignKey:HandledBy;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"` // Wajib menjaga relasi
	CreatedAt time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}
