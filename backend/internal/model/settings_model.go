package model

import "time"

type Settings struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	APIKey       string `json:"apiKey"`
	DownloadPath string `json:"downloadPath"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UpdateSettingsRequest struct {
	APIKey string `json:"apiKey"`
}
