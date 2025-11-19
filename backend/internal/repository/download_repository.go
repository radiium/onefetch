package repository

import (
	"dlbackend/internal/database"
	"dlbackend/internal/model"
)

type DownloadRepository interface {
	List(status []model.DownloadStatus, downloadTypes []model.DownloadType, page, limit int) ([]model.Download, int64, error)
	Create(download *model.Download) error
	GetByID(id string) (*model.Download, error)
	Update(download *model.Download) error
	GetActive() ([]model.Download, error)
	Delete(id string) error
}

type downloadRepository struct {
	db *database.Database
}

func NewDownloadRepository(db *database.Database) DownloadRepository {
	return &downloadRepository{db: db}
}

func (r *downloadRepository) List(status []model.DownloadStatus, downloadTypes []model.DownloadType, page, limit int) ([]model.Download, int64, error) {
	var downloads []model.Download
	var total int64

	query := r.db.Where("is_archived = ?", false)

	if len(status) > 0 {
		query = query.Where("status IN ?", status)
	}
	if len(downloadTypes) > 0 {
		query = query.Where("type IN ?", downloadTypes)
	}

	err := query.Model(&model.Download{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&downloads).Error

	return downloads, total, err
}

func (r *downloadRepository) GetActive() ([]model.Download, error) {
	var downloads []model.Download
	err := r.db.Where("status IN ?", []model.DownloadStatus{
		model.StatusPending,
		model.StatusRequestingInfos,
		model.StatusRequestingToken,
		model.StatusDownloading,
	}).Find(&downloads).Error
	return downloads, err
}

func (r *downloadRepository) Create(download *model.Download) error {
	return r.db.Create(download).Error
}

func (r *downloadRepository) GetByID(id string) (*model.Download, error) {
	var download model.Download
	err := r.db.Where("id = ?", id).First(&download).Error
	return &download, err
}

func (r *downloadRepository) Update(download *model.Download) error {
	return r.db.Save(download).Error
}

func (r *downloadRepository) Delete(id string) error {
	return r.db.Delete(&model.Download{}, "id = ?", id).Error
}
