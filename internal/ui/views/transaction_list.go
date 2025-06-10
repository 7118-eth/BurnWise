package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"budget-tracker/internal/service"
)

type TransactionList struct {
	width  int
	height int
	txService *service.TransactionService
	categoryService *service.CategoryService
}

func NewTransactionList(txService *service.TransactionService, categoryService *service.CategoryService) *TransactionList {
	return &TransactionList{
		txService: txService,
		categoryService: categoryService,
	}
}

func (t *TransactionList) Init() tea.Cmd {
	return nil
}

func (t *TransactionList) Update(msg tea.Msg) (*TransactionList, tea.Cmd) {
	return t, nil
}

func (t *TransactionList) View() string {
	return "Transaction List - Press ESC to go back"
}

func (t *TransactionList) SetSize(width, height int) {
	t.width = width
	t.height = height
}

func (t *TransactionList) HasTransactions() bool {
	return true
}