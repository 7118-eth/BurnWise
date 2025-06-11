package db

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"budget-tracker/internal/models"
)

func InitDB(dbPath string) (*gorm.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	if err := seedDefaultData(db); err != nil {
		return nil, fmt.Errorf("failed to seed default data: %w", err)
	}

	return db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Transaction{},
		&models.Category{},
		&models.Budget{},
		&models.CategoryHistory{},
		&models.RecurringTransaction{},
		&models.RecurringTransactionOccurrence{},
	)
}

func seedDefaultData(db *gorm.DB) error {
	var count int64
	db.Model(&models.Category{}).Where("is_default = ?", true).Count(&count)
	if count > 0 {
		return nil
	}

	categories := models.GetDefaultCategories()
	for _, category := range categories {
		if err := db.Create(&category).Error; err != nil {
			return fmt.Errorf("failed to create default category %s: %w", category.Name, err)
		}
	}

	return nil
}

func GetDefaultDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "data/budget.db"
	}

	return filepath.Join(homeDir, ".local", "share", "budget-tracker", "budget.db")
}