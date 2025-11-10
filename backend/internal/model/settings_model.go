package model

import "time"

type Settings struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	APIKey1fichier string `json:"apiKey1fichier"`
	APIKeyJellyfin string `json:"apiKeyJellyfin"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UpdateSettingsRequest struct {
	APIKey1fichier string `json:"apiKey1fichier"`
	APIKeyJellyfin string `json:"apiKeyJellyfin"`
}
