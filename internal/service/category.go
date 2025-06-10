package service

import (
	"fmt"
	"time"

	"budget-tracker/internal/models"
	"budget-tracker/internal/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(category *models.Category) error {
	if err := category.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	existing, _ := s.repo.FindByName(category.Name, category.Type)
	if existing != nil {
		return fmt.Errorf("category with name '%s' already exists for type %s", category.Name, category.Type)
	}

	return s.repo.Create(category)
}

func (s *CategoryService) Update(category *models.Category) error {
	if err := category.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	existing, _ := s.repo.FindByName(category.Name, category.Type)
	if existing != nil && existing.ID != category.ID {
		return fmt.Errorf("category with name '%s' already exists for type %s", category.Name, category.Type)
	}

	return s.repo.Update(category)
}

func (s *CategoryService) Delete(id uint) error {
	category, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	if category.IsDefault {
		return fmt.Errorf("cannot delete default category")
	}

	count, err := s.repo.GetUsageCount(id)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("cannot delete category with %d transactions", count)
	}

	return s.repo.Delete(id)
}

func (s *CategoryService) GetByID(id uint) (*models.Category, error) {
	return s.repo.GetByID(id)
}

func (s *CategoryService) GetAll() ([]*models.Category, error) {
	return s.repo.GetAll()
}

func (s *CategoryService) GetByType(txType models.TransactionType) ([]*models.Category, error) {
	return s.repo.GetByType(txType)
}

func (s *CategoryService) GetDefault() ([]*models.Category, error) {
	return s.repo.GetDefault()
}

func (s *CategoryService) GetWithTotals(start, end time.Time) ([]*models.CategoryWithTotal, error) {
	return s.repo.GetWithTotals(start, end)
}

func (s *CategoryService) GetCurrentMonthTotals() ([]*models.CategoryWithTotal, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	
	return s.repo.GetWithTotals(start, end)
}

func (s *CategoryService) EnsureDefaultCategories() error {
	defaults := models.GetDefaultCategories()
	
	for _, defaultCat := range defaults {
		existing, _ := s.repo.FindByName(defaultCat.Name, defaultCat.Type)
		if existing == nil {
			category := defaultCat
			if err := s.repo.Create(&category); err != nil {
				return fmt.Errorf("failed to create default category %s: %w", category.Name, err)
			}
		}
	}
	
	return nil
}