package styles

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)

var (
	Primary   = lipgloss.Color("#00BCD4")
	Secondary = lipgloss.Color("#FF5722")
	Success   = lipgloss.Color("#4CAF50")
	Warning   = lipgloss.Color("#FF9800")
	Error     = lipgloss.Color("#F44336")
	Muted     = lipgloss.Color("#9E9E9E")
	
	Income    = lipgloss.Color("#4CAF50")
	Expense   = lipgloss.Color("#F44336")
	
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		Padding(0, 1)
	
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(Primary)
	
	IncomeStyle = lipgloss.NewStyle().
		Foreground(Income).
		Bold(true)
	
	ExpenseStyle = lipgloss.NewStyle().
		Foreground(Expense).
		Bold(true)
	
	BalanceStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary)
	
	SelectedStyle = lipgloss.NewStyle().
		Background(Primary).
		Foreground(lipgloss.Color("#FFFFFF"))
	
	FormLabelStyle = lipgloss.NewStyle().
		Bold(true).
		Width(12)
	
	FormInputStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Muted).
		Padding(0, 1)
	
	FormInputFocusedStyle = FormInputStyle.Copy().
		BorderForeground(Primary)
	
	ButtonStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Background(Primary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	
	ButtonInactiveStyle = ButtonStyle.Copy().
		Background(Muted)
	
	ErrorStyle = lipgloss.NewStyle().
		Foreground(Error).
		Bold(true)
	
	SuccessStyle = lipgloss.NewStyle().
		Foreground(Success).
		Bold(true)
	
	WarningStyle = lipgloss.NewStyle().
		Foreground(Warning).
		Bold(true)
	
	HelpStyle = lipgloss.NewStyle().
		Foreground(Muted).
		Italic(true)
	
	TableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(Primary)
	
	ProgressBarStyle = lipgloss.NewStyle().
		Foreground(Primary)
	
	ProgressBarEmptyStyle = lipgloss.NewStyle().
		Foreground(Muted)
)

func FormatAmount(amount float64, currency string) string {
	prefix := ""
	style := BalanceStyle
	
	if amount < 0 {
		prefix = "-"
		amount = -amount
		style = ExpenseStyle
	} else if amount > 0 {
		prefix = "+"
		style = IncomeStyle
	}
	
	return style.Render(prefix + currency + " " + FormatNumber(amount))
}

func FormatNumber(n float64) string {
	return lipgloss.NewStyle().Render(fmt.Sprintf("%.2f", n))
}

func ProgressBar(percent float64, width int) string {
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}
	
	filled := int(float64(width) * percent / 100)
	empty := width - filled
	
	bar := ProgressBarStyle.Render(strings.Repeat("█", filled))
	bar += ProgressBarEmptyStyle.Render(strings.Repeat("░", empty))
	
	color := Success
	if percent > 80 {
		color = Warning
	}
	if percent > 100 {
		color = Error
	}
	
	return lipgloss.NewStyle().Foreground(color).Render(bar)
}