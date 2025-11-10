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

func (r *settingsRepository) Get() (*model.Settings, error) {
	var settings model.Settings
	err := r.db.First(&settings).Error
	return &settings, err
}

func (r *settingsRepository) Update(settings *model.UpdateSettingsRequest) error {
	return r.db.Model(&model.Settings{}).Where("id = ?", 1).Updates(settings).Error
}
