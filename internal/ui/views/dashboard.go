package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"budget-tracker/internal/models"
	"budget-tracker/internal/service"
	"budget-tracker/internal/ui/styles"
)

type Dashboard struct {
	width    int
	height   int
	
	txService     *service.TransactionService
	budgetService *service.BudgetService
	
	summary      *models.TransactionSummary
	transactions []*models.Transaction
	budgets      []*models.BudgetStatus
	
	loading      bool
	err          error
}

func NewDashboard(txService *service.TransactionService, budgetService *service.BudgetService) *Dashboard {
	return &Dashboard{
		txService:     txService,
		budgetService: budgetService,
		loading:       true,
	}
}

func (d *Dashboard) Init() tea.Cmd {
	return d.loadData
}

func (d *Dashboard) Update(msg tea.Msg) (*Dashboard, tea.Cmd) {
	switch msg := msg.(type) {
	case dashboardDataMsg:
		d.loading = false
		d.summary = msg.summary
		d.transactions = msg.transactions
		d.budgets = msg.budgets
		d.err = msg.err
	}
	
	return d, nil
}

func (d *Dashboard) View() string {
	if d.loading {
		return styles.TitleStyle.Render("Loading...")
	}
	
	if d.err != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", d.err))
	}
	
	header := d.renderHeader()
	summary := d.renderSummary()
	transactions := d.renderRecentTransactions()
	budgets := d.renderBudgetOverview()
	help := d.renderHelp()
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		summary,
		"",
		budgets,
		"",
		transactions,
		"",
		help,
	)
	
	return lipgloss.NewStyle().
		Width(d.width).
		Height(d.height).
		Render(content)
}

func (d *Dashboard) SetSize(width, height int) {
	d.width = width
	d.height = height
}

func (d *Dashboard) renderHeader() string {
	now := time.Now()
	month := now.Format("January 2006")
	
	title := styles.TitleStyle.Render("💰 Budget Tracker")
	date := lipgloss.NewStyle().
		Foreground(styles.Muted).
		Render(month)
	
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		lipgloss.NewStyle().Width(d.width - lipgloss.Width(title) - lipgloss.Width(date)).Render(""),
		date,
	)
	
	return header
}

func (d *Dashboard) renderSummary() string {
	if d.summary == nil {
		return ""
	}
	
	incomeBar := d.renderProgressBar("Income", d.summary.TotalIncome, d.summary.TotalIncome, styles.Income)
	expenseBar := d.renderProgressBar("Expenses", d.summary.TotalExpenses, d.summary.TotalIncome, styles.Expense)
	
	divider := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Render(strings.Repeat("━", d.width-4))
	
	balance := lipgloss.NewStyle().
		Bold(true).
		Render(fmt.Sprintf("Balance:   %s", styles.FormatAmount(d.summary.Balance, "$")))
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		incomeBar,
		expenseBar,
		divider,
		balance,
	)
}

func (d *Dashboard) renderProgressBar(label string, value, max float64, color lipgloss.Color) string {
	if max == 0 {
		max = 1
	}
	
	percent := (value / max) * 100
	if percent > 100 {
		percent = 100
	}
	
	labelStyle := lipgloss.NewStyle().
		Width(10).
		Render(label + ":")
	
	valueStyle := lipgloss.NewStyle().
		Width(12).
		Align(lipgloss.Right).
		Foreground(color).
		Render(fmt.Sprintf("$%.2f", value))
	
	barWidth := d.width - 10 - 12 - 8 - 6
	bar := styles.ProgressBar(percent, barWidth)
	
	percentStyle := lipgloss.NewStyle().
		Width(6).
		Align(lipgloss.Right).
		Render(fmt.Sprintf("%.0f%%", percent))
	
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		labelStyle,
		valueStyle,
		"  ",
		bar,
		"  ",
		percentStyle,
	)
}

func (d *Dashboard) renderRecentTransactions() string {
	if len(d.transactions) == 0 {
		return lipgloss.NewStyle().
			Foreground(styles.Muted).
			Render("No recent transactions")
	}
	
	title := lipgloss.NewStyle().
		Bold(true).
		Render("Recent Transactions")
	
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(12).Render("Date"),
		lipgloss.NewStyle().Width(20).Render("Category"),
		lipgloss.NewStyle().Width(30).Render("Description"),
		lipgloss.NewStyle().Width(12).Align(lipgloss.Right).Render("Amount"),
	)
	
	divider := strings.Repeat("─", d.width-4)
	
	var rows []string
	for i, tx := range d.transactions {
		if i >= 5 {
			break
		}
		
		date := tx.Date.Format("01/02")
		category := fmt.Sprintf("%s %s", tx.Category.Icon, tx.Category.Name)
		description := tx.Description
		if len(description) > 28 {
			description = description[:28] + "..."
		}
		
		amount := styles.FormatAmount(tx.Amount, "$")
		if tx.Type == models.TransactionTypeExpense {
			amount = styles.FormatAmount(-tx.Amount, "$")
		}
		
		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(12).Render(date),
			lipgloss.NewStyle().Width(20).Render(category),
			lipgloss.NewStyle().Width(30).Render(description),
			lipgloss.NewStyle().Width(12).Align(lipgloss.Right).Render(amount),
		)
		
		rows = append(rows, row)
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		header,
		divider,
		strings.Join(rows, "\n"),
	)
}

func (d *Dashboard) renderBudgetOverview() string {
	if len(d.budgets) == 0 {
		return ""
	}
	
	title := lipgloss.NewStyle().
		Bold(true).
		Render("Budget Overview")
	
	var rows []string
	for _, status := range d.budgets {
		if status.Budget.Period != models.BudgetPeriodMonthly {
			continue
		}
		
		category := fmt.Sprintf("%s %s", status.Budget.Category.Icon, status.Budget.Category.Name)
		
		barWidth := 20
		bar := styles.ProgressBar(status.PercentUsed, barWidth)
		
		spent := fmt.Sprintf("$%.0f/$%.0f", status.Spent, status.Budget.Amount)
		
		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(25).Render(category),
			bar,
			"  ",
			lipgloss.NewStyle().Width(15).Align(lipgloss.Right).Render(spent),
		)
		
		rows = append(rows, row)
	}
	
	if len(rows) == 0 {
		return ""
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		strings.Join(rows, "\n"),
	)
}

func (d *Dashboard) renderHelp() string {
	help := []string{
		"[n]ew",
		"[t]ransactions",
		"[b]udgets",
		"[r]eports",
		"[c]ategories",
		"[s]ubscriptions",
		"c[u]rrencies",
		"[q]uit",
	}
	
	return styles.HelpStyle.Render(strings.Join(help, "  "))
}

func (d *Dashboard) loadData() tea.Msg {
	summary, err := d.txService.GetCurrentMonthSummary()
	if err != nil {
		return dashboardDataMsg{err: err}
	}
	
	transactions, err := d.txService.GetRecentTransactions(10)
	if err != nil {
		return dashboardDataMsg{err: err}
	}
	
	budgets, err := d.budgetService.GetAllStatuses()
	if err != nil {
		return dashboardDataMsg{err: err}
	}
	
	return dashboardDataMsg{
		summary:      summary,
		transactions: transactions,
		budgets:      budgets,
	}
}

type dashboardDataMsg struct {
	summary      *models.TransactionSummary
	transactions []*models.Transaction
	budgets      []*models.BudgetStatus
	err          error
}