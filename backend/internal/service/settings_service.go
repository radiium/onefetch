package service

import (
	"dlbackend/internal/errors"
	"dlbackend/internal/model"
	"dlbackend/internal/repository"
	"fmt"
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
	settings, err := ss.repo.Get()
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to retrieve settings: %v", err))
	}

	return settings, nil
}

func (ss *settingsService) UpdateSettings(settings *model.UpdateSettingsRequest) (*model.Settings, error) {
	if err := ss.repo.Update(settings); err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to update settings: %v", err))
	}

	updated, err := ss.repo.Get()
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to retrieve settings: %v", err))
	}

	return updated, nil
}
