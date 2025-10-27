package database

import (
	"dlbackend/internal/model"
	"dlbackend/pkg/config"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	*gorm.DB
}

func New(cfg *config.Config) (*Database, error) {
	var err error

	dbPath := filepath.Join(cfg.DataPath, "onefetch.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&model.Settings{}, &model.Download{})
	if err != nil {
		return nil, err
	}

	// Initialize default settings if not exists
	var count int64
	db.Model(&model.Settings{}).Count(&count)
	if count == 0 {
		db.Create(&model.Settings{
			APIKey:       "",
			DownloadPath: cfg.DLPath,
		})
	}

	return &Database{db}, err
}

func (db *Database) Close() error {
	var err error
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}
