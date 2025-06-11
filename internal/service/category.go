package service

import (
	"fmt"
	"time"

	"burnwise/internal/models"
	"burnwise/internal/repository"
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

	// Get the old category to track changes
	oldCategory, err := s.repo.GetByID(category.ID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Check for duplicate names
	existing, _ := s.repo.FindByName(category.Name, category.Type)
	if existing != nil && existing.ID != category.ID {
		return fmt.Errorf("category with name '%s' already exists for type %s", category.Name, category.Type)
	}

	// Update the category
	if err := s.repo.Update(category); err != nil {
		return err
	}

	// Track history if there were changes
	if oldCategory.Name != category.Name || oldCategory.Icon != category.Icon || oldCategory.Color != category.Color {
		history := &models.CategoryHistory{
			CategoryID: category.ID,
			Action:     models.CategoryActionEdited,
		}

		if oldCategory.Name != category.Name {
			history.OldName = oldCategory.Name
			history.NewName = category.Name
		}
		if oldCategory.Icon != category.Icon {
			history.OldIcon = oldCategory.Icon
			history.NewIcon = category.Icon
		}
		if oldCategory.Color != category.Color {
			history.OldColor = oldCategory.Color
			history.NewColor = category.Color
		}

		if err := s.repo.CreateHistory(history); err != nil {
			// Log error but don't fail the update
			fmt.Printf("Warning: failed to record category history: %v\n", err)
		}
	}

	return nil
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

func (s *CategoryService) MergeCategories(sourceID, targetID uint) error {
	if sourceID == targetID {
		return fmt.Errorf("cannot merge a category with itself")
	}

	// Verify source category
	source, err := s.repo.GetByID(sourceID)
	if err != nil {
		return fmt.Errorf("source category not found: %w", err)
	}

	// Verify target category
	target, err := s.repo.GetByID(targetID)
	if err != nil {
		return fmt.Errorf("target category not found: %w", err)
	}

	// Ensure both categories are of the same type
	if source.Type != target.Type {
		return fmt.Errorf("cannot merge categories of different types (%s -> %s)", source.Type, target.Type)
	}

	// Prevent merging default categories
	if source.IsDefault {
		return fmt.Errorf("cannot merge default category '%s'", source.Name)
	}

	return s.repo.MergeCategories(sourceID, targetID)
}

func (s *CategoryService) GetAllWithUsageCount() ([]*models.CategoryWithTotal, error) {
	return s.repo.GetAllWithUsageCount()
}

func (s *CategoryService) GetHistory(categoryID uint) ([]*models.CategoryHistory, error) {
	return s.repo.GetHistory(categoryID)
}

func (s *CategoryService) GetUsageCount(categoryID uint) (int64, error) {
	return s.repo.GetUsageCount(categoryID)
}