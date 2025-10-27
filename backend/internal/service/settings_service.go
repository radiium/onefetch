package service

import (
	"dlbackend/internal/model"
	"dlbackend/internal/repository"
)

type SettingsService interface {
	GetSettings() (*model.Settings, error)
	UpdateSettings(settings *model.UpdateSettingsRequest) (*model.Settings, error)
}

type settingsService struct {
	repo repository.SettingsRepository
}

func NewSettingsService(repo repository.SettingsRepository) SettingsService {
	return &settingsService{repo: repo}
}

func (ss *settingsService) GetSettings() (*model.Settings, error) {
	return ss.repo.Get()
}

func (ss *settingsService) UpdateSettings(settings *model.UpdateSettingsRequest) (*model.Settings, error) {
	if err := ss.repo.Update(settings); err != nil {
		return nil, err
	}
	return ss.repo.Get()
}
