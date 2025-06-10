package repository

import (
	"time"
	
	"gorm.io/gorm"

	"budget-tracker/internal/models"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

func (r *CategoryRepository) GetByID(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *CategoryRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Transaction{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		
		if count > 0 {
			return gorm.ErrRecordNotFound
		}
		
		return tx.Delete(&models.Category{}, id).Error
	})
}

func (r *CategoryRepository) GetAll() ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.Order("type ASC, name ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) GetByType(txType models.TransactionType) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.Where("type = ?", txType).Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) GetDefault() ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.Where("is_default = ?", true).Order("type ASC, name ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) GetWithTotals(start, end time.Time) ([]*models.CategoryWithTotal, error) {
	var results []*models.CategoryWithTotal

	err := r.db.Table("categories").
		Select("categories.*, COALESCE(SUM(transactions.amount_usd), 0) as total, COUNT(transactions.id) as count").
		Joins("LEFT JOIN transactions ON categories.id = transactions.category_id AND transactions.date >= ? AND transactions.date <= ? AND transactions.deleted_at IS NULL", start, end).
		Where("categories.deleted_at IS NULL").
		Group("categories.id").
		Order("categories.type ASC, total DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	totalsByType := make(map[models.TransactionType]float64)
	for _, result := range results {
		totalsByType[result.Type] += result.Total
	}

	for _, result := range results {
		if total := totalsByType[result.Type]; total > 0 {
			result.Percentage = (result.Total / total) * 100
		}
	}

	return results, nil
}

func (r *CategoryRepository) FindByName(name string, txType models.TransactionType) (*models.Category, error) {
	var category models.Category
	err := r.db.Where("name = ? AND type = ?", name, txType).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) GetUsageCount(categoryID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Transaction{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}