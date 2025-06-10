package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"budget-tracker/internal/models"
	"budget-tracker/internal/service"
	"budget-tracker/internal/ui/styles"
)

type TransactionList struct {
	width           int
	height          int
	txService       *service.TransactionService
	categoryService *service.CategoryService
	
	transactions    []*models.Transaction
	table           table.Model
	loading         bool
	err             error
	
	filter          *models.TransactionFilter
	showFilter      bool
}

type transactionDeletedMsg struct{}
type TransactionEditMsg struct{ Transaction *models.Transaction }

func NewTransactionList(txService *service.TransactionService, categoryService *service.CategoryService) *TransactionList {
	columns := []table.Column{
		{Title: "Date", Width: 10},
		{Title: "Type", Width: 8},
		{Title: "Category", Width: 20},
		{Title: "Description", Width: 30},
		{Title: "Amount", Width: 12},
		{Title: "Currency", Width: 8},
	}
	
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.Primary).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(styles.Primary).
		Bold(false)
	t.SetStyles(s)
	
	return &TransactionList{
		txService:       txService,
		categoryService: categoryService,
		table:           t,
		filter:          &models.TransactionFilter{},
	}
}

func (t *TransactionList) Init() tea.Cmd {
	t.loading = true
	return t.loadTransactions
}

func (t *TransactionList) Update(msg tea.Msg) (*TransactionList, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.SetSize(msg.Width, msg.Height)
		
	case tea.KeyMsg:
		if t.showFilter {
			return t.handleFilterKeys(msg)
		}
		
		switch msg.String() {
		case "e":
			if len(t.transactions) > 0 {
				selected := t.table.SelectedRow()
				if selected != nil && len(selected) > 0 {
					idx := t.table.Cursor()
					if idx < len(t.transactions) {
						return t, func() tea.Msg { 
							return TransactionEditMsg{Transaction: t.transactions[idx]}
						}
					}
				}
			}
		case "d":
			if len(t.transactions) > 0 {
				idx := t.table.Cursor()
				if idx < len(t.transactions) {
					return t, t.deleteTransaction(t.transactions[idx].ID)
				}
			}
		case "f":
			t.showFilter = !t.showFilter
		case "/":
			t.showFilter = true
		}
		
	case transactionsLoadedMsg:
		t.loading = false
		t.transactions = msg.transactions
		t.err = msg.err
		t.updateTable()
		
	case transactionDeletedMsg:
		return t, t.loadTransactions
	}
	
	if !t.showFilter {
		t.table, cmd = t.table.Update(msg)
	}
	
	return t, cmd
}

func (t *TransactionList) View() string {
	if t.loading {
		return styles.TitleStyle.Render("Loading transactions...")
	}
	
	if t.err != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", t.err))
	}
	
	header := t.renderHeader()
	
	var content string
	if len(t.transactions) == 0 {
		content = lipgloss.NewStyle().
			Foreground(styles.Muted).
			Padding(2).
			Render("No transactions found. Press 'n' to add a new transaction.")
	} else {
		content = t.table.View()
	}
	
	help := t.renderHelp()
	
	if t.showFilter {
		content += "\n\n" + t.renderFilter()
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		help,
	)
}

func (t *TransactionList) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.table.SetHeight(height - 10)
	t.table.SetWidth(width)
}

func (t *TransactionList) HasTransactions() bool {
	return len(t.transactions) > 0
}

func (t *TransactionList) renderHeader() string {
	title := styles.TitleStyle.Render("💰 All Transactions")
	
	count := fmt.Sprintf("%d transactions", len(t.transactions))
	countStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		lipgloss.NewStyle().Width(t.width - lipgloss.Width(title) - lipgloss.Width(count) - 2).Render(""),
		countStyle.Render(count),
	)
}

func (t *TransactionList) renderHelp() string {
	help := []string{
		"[n]ew",
		"[e]dit",
		"[d]elete",
		"[f]ilter",
		"[/]search",
		"[esc]back",
	}
	
	return styles.HelpStyle.Render(strings.Join(help, "  "))
}

func (t *TransactionList) renderFilter() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1).
		Render("Filter options coming soon... Press 'f' to hide")
}

func (t *TransactionList) updateTable() {
	rows := []table.Row{}
	
	for _, tx := range t.transactions {
		date := tx.Date.Format("2006-01-02")
		txType := string(tx.Type)
		category := fmt.Sprintf("%s %s", tx.Category.Icon, tx.Category.Name)
		description := tx.Description
		if len(description) > 28 {
			description = description[:28] + "..."
		}
		
		amount := fmt.Sprintf("%.2f", tx.Amount)
		if tx.Type == models.TransactionTypeExpense {
			amount = "-" + amount
		} else if tx.Type == models.TransactionTypeIncome {
			amount = "+" + amount
		}
		
		row := table.Row{date, txType, category, description, amount, tx.Currency}
		rows = append(rows, row)
	}
	
	t.table.SetRows(rows)
}

func (t *TransactionList) loadTransactions() tea.Msg {
	transactions, err := t.txService.GetByFilter(t.filter)
	return transactionsLoadedMsg{
		transactions: transactions,
		err:          err,
	}
}

func (t *TransactionList) deleteTransaction(id uint) tea.Cmd {
	return func() tea.Msg {
		if err := t.txService.Delete(id); err != nil {
			return errMsg{err}
		}
		return transactionDeletedMsg{}
	}
}

func (t *TransactionList) handleFilterKeys(msg tea.KeyMsg) (*TransactionList, tea.Cmd) {
	switch msg.String() {
	case "esc", "f":
		t.showFilter = false
	}
	return t, nil
}

type transactionsLoadedMsg struct {
	transactions []*models.Transaction
	err          error
}

type errMsg struct{ error }