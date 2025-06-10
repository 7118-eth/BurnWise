package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"budget-tracker/internal/service"
)

type BudgetForm struct {
	width  int
	height int
	budgetService *service.BudgetService
	categoryService *service.CategoryService
}

type BudgetSavedMsg struct{}
type BudgetCancelledMsg struct{}

func NewBudgetForm(budgetService *service.BudgetService, categoryService *service.CategoryService) *BudgetForm {
	return &BudgetForm{
		budgetService: budgetService,
		categoryService: categoryService,
	}
}

func (b *BudgetForm) Init() tea.Cmd {
	return nil
}

func (b *BudgetForm) Update(msg tea.Msg) (*BudgetForm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return b, func() tea.Msg { return BudgetCancelledMsg{} }
		}
	}
	return b, nil
}

func (b *BudgetForm) View() string {
	return "Budget Form - Press ESC to cancel"
}

func (b *BudgetForm) SetSize(width, height int) {
	b.width = width
	b.height = height
}

func (b *BudgetForm) Reset() {
	// Reset form fields
}