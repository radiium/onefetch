package database

import (
	"dlbackend/internal/config"
	"dlbackend/internal/model"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database wraps gorm.DB to provide database operations.
type Database struct {
	*gorm.DB
}

// New creates a Database instance, runs migrations, and initializes default settings.
func New() (*Database, error) {
	var err error

	dbPath := filepath.Join(config.Cfg.DataPath, "onefetch.db")
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
			APIKey1fichier: "",
			APIKeyJellyfin: "",
		})
	}

	return &Database{db}, err
}

// Close closes the database connection.
func (db *Database) Close() error {
	var err error
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}
