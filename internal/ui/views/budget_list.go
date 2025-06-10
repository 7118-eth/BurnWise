package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"budget-tracker/internal/service"
)

type BudgetList struct {
	width  int
	height int
	budgetService *service.BudgetService
	categoryService *service.CategoryService
}

func NewBudgetList(budgetService *service.BudgetService, categoryService *service.CategoryService) *BudgetList {
	return &BudgetList{
		budgetService: budgetService,
		categoryService: categoryService,
	}
}

func (b *BudgetList) Init() tea.Cmd {
	return nil
}

func (b *BudgetList) Update(msg tea.Msg) (*BudgetList, tea.Cmd) {
	return b, nil
}

func (b *BudgetList) View() string {
	return "Budget List - Press ESC to go back, N to create new budget"
}

func (b *BudgetList) SetSize(width, height int) {
	b.width = width
	b.height = height
}