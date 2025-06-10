package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"budget-tracker/internal/service"
)

type Reports struct {
	width  int
	height int
	txService *service.TransactionService
	categoryService *service.CategoryService
	budgetService *service.BudgetService
}

func NewReports(txService *service.TransactionService, categoryService *service.CategoryService, budgetService *service.BudgetService) *Reports {
	return &Reports{
		txService: txService,
		categoryService: categoryService,
		budgetService: budgetService,
	}
}

func (r *Reports) Init() tea.Cmd {
	return nil
}

func (r *Reports) Update(msg tea.Msg) (*Reports, tea.Cmd) {
	return r, nil
}

func (r *Reports) View() string {
	return "Reports - Press ESC to go back"
}

func (r *Reports) SetSize(width, height int) {
	r.width = width
	r.height = height
}