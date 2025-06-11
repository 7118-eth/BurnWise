package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"burnwise/internal/models"
	"burnwise/internal/service"
	"burnwise/internal/ui/styles"
)

type Reports struct {
	width           int
	height          int
	txService       *service.TransactionService
	categoryService *service.CategoryService
	budgetService   *service.BudgetService
	
	monthSummary    *models.TransactionSummary
	yearSummary     *models.TransactionSummary
	categoryTotals  []*models.CategoryWithTotal
	budgetStatuses  []*models.BudgetStatus
	
	selectedMonth   time.Month
	selectedYear    int
	loading         bool
	err             error
}

func NewReports(txService *service.TransactionService, categoryService *service.CategoryService, budgetService *service.BudgetService) *Reports {
	now := time.Now()
	return &Reports{
		txService:       txService,
		categoryService: categoryService,
		budgetService:   budgetService,
		selectedMonth:   now.Month(),
		selectedYear:    now.Year(),
	}
}

func (r *Reports) Init() tea.Cmd {
	r.loading = true
	return r.loadReportData
}

func (r *Reports) Update(msg tea.Msg) (*Reports, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.SetSize(msg.Width, msg.Height)
		
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			r.selectedMonth--
			if r.selectedMonth < 1 {
				r.selectedMonth = 12
				r.selectedYear--
			}
			return r, r.loadReportData
		case "right":
			r.selectedMonth++
			if r.selectedMonth > 12 {
				r.selectedMonth = 1
				r.selectedYear++
			}
			return r, r.loadReportData
		}
		
	case reportDataMsg:
		r.loading = false
		r.monthSummary = msg.monthSummary
		r.yearSummary = msg.yearSummary
		r.categoryTotals = msg.categoryTotals
		r.budgetStatuses = msg.budgetStatuses
		r.err = msg.err
	}
	
	return r, nil
}

func (r *Reports) View() string {
	if r.loading {
		return styles.TitleStyle.Render("Loading reports...")
	}
	
	if r.err != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", r.err))
	}
	
	header := r.renderHeader()
	monthSummary := r.renderMonthSummary()
	yearSummary := r.renderYearSummary()
	categoryBreakdown := r.renderCategoryBreakdown()
	budgetPerformance := r.renderBudgetPerformance()
	help := r.renderHelp()
	
	leftColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		monthSummary,
		"",
		yearSummary,
	)
	
	rightColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		categoryBreakdown,
		"",
		budgetPerformance,
	)
	
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(r.width/2).Render(leftColumn),
		lipgloss.NewStyle().Width(r.width/2).Render(rightColumn),
	)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		help,
	)
}

func (r *Reports) SetSize(width, height int) {
	r.width = width
	r.height = height
}

func (r *Reports) renderHeader() string {
	title := styles.TitleStyle.Render("📊 Financial Reports")
	
	monthNav := fmt.Sprintf("← %s %d →", r.selectedMonth.String(), r.selectedYear)
	navStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true)
	
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		lipgloss.NewStyle().Width(r.width - lipgloss.Width(title) - lipgloss.Width(monthNav) - 2).Render(""),
		navStyle.Render(monthNav),
	)
}

func (r *Reports) renderMonthSummary() string {
	if r.monthSummary == nil {
		return ""
	}
	
	title := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Render(fmt.Sprintf("%s %d Summary", r.selectedMonth.String(), r.selectedYear))
	
	income := styles.IncomeStyle.Render(fmt.Sprintf("Income:    $%.2f", r.monthSummary.TotalIncome))
	expenses := styles.ExpenseStyle.Render(fmt.Sprintf("Expenses:  $%.2f", r.monthSummary.TotalExpenses))
	
	balanceStyle := styles.BalanceStyle
	if r.monthSummary.Balance < 0 {
		balanceStyle = styles.ExpenseStyle
	}
	balance := balanceStyle.Render(fmt.Sprintf("Balance:   $%.2f", r.monthSummary.Balance))
	
	divider := strings.Repeat("─", 25)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		income,
		expenses,
		divider,
		balance,
	)
}

func (r *Reports) renderYearSummary() string {
	if r.yearSummary == nil {
		return ""
	}
	
	title := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Render(fmt.Sprintf("%d Year-to-Date", r.selectedYear))
	
	income := styles.IncomeStyle.Render(fmt.Sprintf("Income:    $%.2f", r.yearSummary.TotalIncome))
	expenses := styles.ExpenseStyle.Render(fmt.Sprintf("Expenses:  $%.2f", r.yearSummary.TotalExpenses))
	
	avgMonthly := r.yearSummary.TotalExpenses / float64(r.selectedMonth)
	avgStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	average := avgStyle.Render(fmt.Sprintf("Avg/Month: $%.2f", avgMonthly))
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		income,
		expenses,
		average,
	)
}

func (r *Reports) renderCategoryBreakdown() string {
	if len(r.categoryTotals) == 0 {
		return ""
	}
	
	title := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Render("Category Breakdown")
	
	var rows []string
	for i, cat := range r.categoryTotals {
		if i >= 8 { // Limit to top 8 categories
			break
		}
		if cat.Total == 0 {
			continue
		}
		
		name := fmt.Sprintf("%s %s", cat.Icon, cat.Name)
		if len(name) > 20 {
			name = name[:20] + "..."
		}
		
		bar := r.renderMiniBar(cat.Percentage, 10)
		amount := fmt.Sprintf("$%.2f", cat.Total)
		
		row := fmt.Sprintf("%-22s %s %8s", name, bar, amount)
		rows = append(rows, row)
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		strings.Join(rows, "\n"),
	)
}

func (r *Reports) renderBudgetPerformance() string {
	if len(r.budgetStatuses) == 0 {
		return ""
	}
	
	title := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Render("Budget Performance")
	
	var rows []string
	var overBudgetCount int
	
	for _, status := range r.budgetStatuses {
		if status.Budget.Period != models.BudgetPeriodMonthly {
			continue
		}
		
		name := fmt.Sprintf("%s %s", status.Budget.Category.Icon, status.Budget.Category.Name)
		if len(name) > 18 {
			name = name[:18] + "..."
		}
		
		percentStyle := styles.SuccessStyle
		if status.PercentUsed > 80 {
			percentStyle = styles.WarningStyle
		}
		if status.PercentUsed > 100 {
			percentStyle = styles.ErrorStyle
			overBudgetCount++
		}
		
		percent := percentStyle.Render(fmt.Sprintf("%.0f%%", status.PercentUsed))
		spent := fmt.Sprintf("$%.0f/$%.0f", status.Spent, status.Budget.Amount)
		
		row := fmt.Sprintf("%-20s %6s %14s", name, percent, spent)
		rows = append(rows, row)
	}
	
	summary := ""
	if overBudgetCount > 0 {
		summary = styles.ErrorStyle.Render(fmt.Sprintf("\n%d budgets exceeded!", overBudgetCount))
	} else {
		summary = styles.SuccessStyle.Render("\nAll budgets on track!")
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		strings.Join(rows, "\n"),
		summary,
	)
}

func (r *Reports) renderMiniBar(percent float64, width int) string {
	if percent > 100 {
		percent = 100
	}
	
	filled := int(float64(width) * percent / 100)
	empty := width - filled
	
	return lipgloss.NewStyle().Foreground(styles.Primary).Render(
		strings.Repeat("█", filled) + strings.Repeat("░", empty),
	)
}

func (r *Reports) renderHelp() string {
	help := []string{
		"[←/→]navigate months",
		"[esc]back",
	}
	
	return styles.HelpStyle.Render(strings.Join(help, "  "))
}

func (r *Reports) loadReportData() tea.Msg {
	monthSummary, err := r.txService.GetMonthSummary(r.selectedYear, r.selectedMonth)
	if err != nil {
		return reportDataMsg{err: err}
	}
	
	yearSummary, err := r.txService.GetYearSummary(r.selectedYear)
	if err != nil {
		return reportDataMsg{err: err}
	}
	
	// Get category totals for the selected month
	start := time.Date(r.selectedYear, r.selectedMonth, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	
	categoryTotals, err := r.txService.GetCategorySummary(start, end)
	if err != nil {
		return reportDataMsg{err: err}
	}
	
	budgetStatuses, err := r.budgetService.GetAllStatuses()
	if err != nil {
		return reportDataMsg{err: err}
	}
	
	return reportDataMsg{
		monthSummary:   monthSummary,
		yearSummary:    yearSummary,
		categoryTotals: categoryTotals,
		budgetStatuses: budgetStatuses,
	}
}

type reportDataMsg struct {
	monthSummary   *models.TransactionSummary
	yearSummary    *models.TransactionSummary
	categoryTotals []*models.CategoryWithTotal
	budgetStatuses []*models.BudgetStatus
	err            error
}