package models

import (
	"time"
)

// User inawakilisha jedwali la watumiaji kwenye database
type User struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	FullName     string    `gorm:"type:varchar(255);not null" json:"full_name"`
	Email        string    `gorm:"type:varchar(255);unique;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"` // "-" inaficha password isiende kwenye API
	Role         string    `gorm:"type:varchar(50);default:'user'" json:"role"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relationships
	Profile      Profile   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"profile,omitempty"`
}

// Profile inawakilisha wasifu wa mtumiaji (Bio, Picha)
type Profile struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;unique;not null" json:"user_id"`
	Bio       string    `gorm:"type:text" json:"bio"`
	AvatarURL string    `gorm:"type:varchar(500)" json:"avatar_url"`
	Website   string    `gorm:"type:varchar(255)" json:"website"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
