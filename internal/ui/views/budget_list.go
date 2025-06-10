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

type BudgetList struct {
	width           int
	height          int
	budgetService   *service.BudgetService
	categoryService *service.CategoryService
	
	budgets         []*models.BudgetStatus
	table           table.Model
	loading         bool
	err             error
}

type budgetDeletedMsg struct{}
type BudgetEditMsg struct{ Budget *models.Budget }

func NewBudgetList(budgetService *service.BudgetService, categoryService *service.CategoryService) *BudgetList {
	columns := []table.Column{
		{Title: "Category", Width: 20},
		{Title: "Period", Width: 10},
		{Title: "Budget", Width: 12},
		{Title: "Spent", Width: 12},
		{Title: "Remaining", Width: 12},
		{Title: "Progress", Width: 20},
		{Title: "Status", Width: 10},
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
	
	return &BudgetList{
		budgetService:   budgetService,
		categoryService: categoryService,
		table:           t,
	}
}

func (b *BudgetList) Init() tea.Cmd {
	b.loading = true
	return b.loadBudgets
}

func (b *BudgetList) Update(msg tea.Msg) (*BudgetList, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.SetSize(msg.Width, msg.Height)
		
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			if len(b.budgets) > 0 {
				idx := b.table.Cursor()
				if idx < len(b.budgets) {
					return b, func() tea.Msg { 
						return BudgetEditMsg{Budget: &b.budgets[idx].Budget}
					}
				}
			}
		case "d":
			if len(b.budgets) > 0 {
				idx := b.table.Cursor()
				if idx < len(b.budgets) {
					return b, b.deleteBudget(b.budgets[idx].Budget.ID)
				}
			}
		}
		
	case budgetsLoadedMsg:
		b.loading = false
		b.budgets = msg.budgets
		b.err = msg.err
		b.updateTable()
		
	case budgetDeletedMsg:
		return b, b.loadBudgets
	}
	
	b.table, cmd = b.table.Update(msg)
	return b, cmd
}

func (b *BudgetList) View() string {
	if b.loading {
		return styles.TitleStyle.Render("Loading budgets...")
	}
	
	if b.err != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", b.err))
	}
	
	header := b.renderHeader()
	
	var content string
	if len(b.budgets) == 0 {
		content = lipgloss.NewStyle().
			Foreground(styles.Muted).
			Padding(2).
			Render("No budgets found. Press 'n' to create a budget.")
	} else {
		content = b.table.View()
	}
	
	help := b.renderHelp()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		help,
	)
}

func (b *BudgetList) SetSize(width, height int) {
	b.width = width
	b.height = height
	b.table.SetHeight(height - 10)
	b.table.SetWidth(width)
}

func (b *BudgetList) renderHeader() string {
	title := styles.TitleStyle.Render("📊 Budget Management")
	
	var totalBudget, totalSpent float64
	for _, status := range b.budgets {
		if status.Budget.Period == models.BudgetPeriodMonthly {
			totalBudget += status.Budget.Amount
			totalSpent += status.Spent
		}
	}
	
	summary := fmt.Sprintf("Monthly: $%.2f / $%.2f", totalSpent, totalBudget)
	summaryStyle := lipgloss.NewStyle().Foreground(styles.Primary)
	
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		lipgloss.NewStyle().Width(b.width - lipgloss.Width(title) - lipgloss.Width(summary) - 2).Render(""),
		summaryStyle.Render(summary),
	)
}

func (b *BudgetList) renderHelp() string {
	help := []string{
		"[n]ew",
		"[e]dit",
		"[d]elete",
		"[esc]back",
	}
	
	return styles.HelpStyle.Render(strings.Join(help, "  "))
}

func (b *BudgetList) updateTable() {
	rows := []table.Row{}
	
	for _, status := range b.budgets {
		category := fmt.Sprintf("%s %s", status.Budget.Category.Icon, status.Budget.Category.Name)
		period := string(status.Budget.Period)
		budget := fmt.Sprintf("$%.2f", status.Budget.Amount)
		spent := fmt.Sprintf("$%.2f", status.Spent)
		remaining := fmt.Sprintf("$%.2f", status.Remaining)
		
		// Progress bar
		progress := ""
		barWidth := 15
		filled := int(float64(barWidth) * status.PercentUsed / 100)
		if filled > barWidth {
			filled = barWidth
		}
		empty := barWidth - filled
		
		progressColor := styles.Success
		if status.PercentUsed > 80 {
			progressColor = styles.Warning
		}
		if status.PercentUsed > 100 {
			progressColor = styles.Error
		}
		
		progress = lipgloss.NewStyle().Foreground(progressColor).Render(
			strings.Repeat("█", filled) + strings.Repeat("░", empty),
		)
		
		statusText := "OK"
		if status.IsOverBudget {
			statusText = "OVER"
		}
		
		row := table.Row{category, period, budget, spent, remaining, progress, statusText}
		rows = append(rows, row)
	}
	
	b.table.SetRows(rows)
}

func (b *BudgetList) loadBudgets() tea.Msg {
	budgets, err := b.budgetService.GetAllStatuses()
	return budgetsLoadedMsg{
		budgets: budgets,
		err:     err,
	}
}

func (b *BudgetList) deleteBudget(id uint) tea.Cmd {
	return func() tea.Msg {
		if err := b.budgetService.Delete(id); err != nil {
			return errMsg{err}
		}
		return budgetDeletedMsg{}
	}
}

type budgetsLoadedMsg struct {
	budgets []*models.BudgetStatus
	err     error
}