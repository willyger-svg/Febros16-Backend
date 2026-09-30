package models

import (
	"time"
)

// User inawakilisha jedwali la watumiaji kwenye database
type User struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FullName     string    `gorm:"type:varchar(255);not null"`
	Email        string    `gorm:"type:varchar(255);unique;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	Role         string    `gorm:"type:varchar(50);default:'user'"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
