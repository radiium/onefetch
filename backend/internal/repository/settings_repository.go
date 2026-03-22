package repository

import (
	"dlbackend/internal/database"
	"dlbackend/internal/model"
)

type SettingsRepository interface {
	Get() (*model.Settings, error)
	Update(settings *model.UpdateSettingsRequest) error
}

type settingsRepository struct {
	db *database.Database
}

func NewSettingsRepository(db *database.Database) SettingsRepository {
	return &settingsRepository{db: db}
}

// Get returns the single settings row.
// NOTE: assumes a settings row with id=1 exists (seeded at startup).
// Returns a GORM "record not found" error if the table is empty.
func (r *settingsRepository) Get() (*model.Settings, error) {
	var settings model.Settings
	err := r.db.First(&settings).Error
	return &settings, err
}

// Update persists settings changes.
// NOTE: targets the row with id=1 (always seeded at startup).
// If no matching row exists, GORM silently updates 0 rows without error.
func (r *settingsRepository) Update(settings *model.UpdateSettingsRequest) error {
	return r.db.Model(&model.Settings{}).Where("id = ?", 1).Updates(settings).Error
}
