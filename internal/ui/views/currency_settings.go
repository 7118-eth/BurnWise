package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"budget-tracker/internal/service"
	"budget-tracker/internal/ui/styles"
)

type currencyItem struct {
	code    string
	enabled bool
}

func (i currencyItem) FilterValue() string { return i.code }

func (i currencyItem) Title() string {
	status := "○"
	if i.enabled {
		status = "●"
	}
	return fmt.Sprintf("%s %s", status, i.code)
}

func (i currencyItem) Description() string {
	if i.enabled {
		return "Enabled"
	}
	return "Disabled"
}

type CurrencySettings struct {
	list            list.Model
	currencies      []currencyItem
	settingsService *service.SettingsService
	currencyService *service.CurrencyService
	txService       *service.TransactionService
	width           int
	height          int
	err             error
	message         string
}

var currencyKeys = struct {
	Toggle key.Binding
	Back   key.Binding
	Enter  key.Binding
}{
	Toggle: key.NewBinding(
		key.WithKeys(" ", "enter"),
		key.WithHelp("space/enter", "toggle"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc/q", "back"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "toggle"),
	),
}

func NewCurrencySettings(settingsService *service.SettingsService, currencyService *service.CurrencyService, txService *service.TransactionService) *CurrencySettings {
	// Get all available currencies
	allCurrencies := currencyService.GetAllAvailableCurrencies()
	enabledCurrencies := settingsService.GetEnabledCurrencies()

	// Create currency items
	items := make([]list.Item, 0, len(allCurrencies))
	currencyItems := make([]currencyItem, 0, len(allCurrencies))

	// Create a map for quick lookup
	enabledMap := make(map[string]bool)
	for _, c := range enabledCurrencies {
		enabledMap[c] = true
	}

	for _, code := range allCurrencies {
		item := currencyItem{
			code:    code,
			enabled: enabledMap[code],
		}
		currencyItems = append(currencyItems, item)
		items = append(items, item)
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Currency Settings"
	l.Styles.Title = styles.TitleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()

	// Add help text
	l.SetShowHelp(true)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			currencyKeys.Toggle,
			currencyKeys.Back,
		}
	}
	l.SetStatusBarItemName("currency", "currencies")
	l.Styles.StatusBar = l.Styles.StatusBar.Foreground(lipgloss.Color("240"))
	l.SetShowHelp(true)
	l.Help.ShowAll = false

	return &CurrencySettings{
		list:            l,
		currencies:      currencyItems,
		settingsService: settingsService,
		currencyService: currencyService,
		txService:       txService,
	}
}

func (m *CurrencySettings) Init() tea.Cmd {
	return nil
}

func (m *CurrencySettings) Update(msg tea.Msg) (*CurrencySettings, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-2)

	case tea.KeyMsg:
		// Clear message on any key press
		if m.message != "" {
			m.message = ""
		}

		switch {
		case key.Matches(msg, currencyKeys.Back):
			return m, func() tea.Msg { return BackToDashboardMsg{} }

		case key.Matches(msg, currencyKeys.Toggle):
			if i, ok := m.list.SelectedItem().(currencyItem); ok {
				// Toggle currency
				if i.enabled {
					// Try to disable
					count, err := m.txService.CountByCurrency(i.code)
					if err != nil {
						m.err = err
						m.message = fmt.Sprintf("Error checking currency usage: %v", err)
					} else if count > 0 {
						m.message = fmt.Sprintf("Cannot disable %s: %d transactions use this currency", i.code, count)
					} else if i.code == m.settingsService.GetDefaultCurrency() {
						m.message = fmt.Sprintf("Cannot disable default currency %s", i.code)
					} else {
						// Disable the currency
						if err := m.settingsService.DisableCurrency(i.code, m.txService); err != nil {
							m.err = err
							m.message = fmt.Sprintf("Failed to disable currency: %v", err)
						} else {
							m.updateCurrencyList()
							m.message = fmt.Sprintf("Disabled %s", i.code)
						}
					}
				} else {
					// Enable the currency
					if err := m.settingsService.EnableCurrency(i.code); err != nil {
						m.err = err
						m.message = fmt.Sprintf("Failed to enable currency: %v", err)
					} else {
						m.updateCurrencyList()
						m.message = fmt.Sprintf("Enabled %s", i.code)
					}
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *CurrencySettings) updateCurrencyList() {
	// Get current enabled currencies
	enabledCurrencies := m.settingsService.GetEnabledCurrencies()
	enabledMap := make(map[string]bool)
	for _, c := range enabledCurrencies {
		enabledMap[c] = true
	}

	// Update currency items
	for i := range m.currencies {
		m.currencies[i].enabled = enabledMap[m.currencies[i].code]
	}

	// Update list items
	items := make([]list.Item, len(m.currencies))
	for i, curr := range m.currencies {
		items[i] = curr
	}
	m.list.SetItems(items)
}

func (m *CurrencySettings) View() string {
	var b strings.Builder

	// Title bar
	titleBar := styles.TitleStyle.Render("Currency Settings")
	b.WriteString(titleBar + "\n\n")

	// Default currency info
	defaultInfo := fmt.Sprintf("Default Currency: %s", lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(m.settingsService.GetDefaultCurrency()))
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(defaultInfo) + "\n\n")

	// List
	b.WriteString(m.list.View())

	// Message or error
	if m.message != "" {
		msgStyle := styles.SuccessStyle
		if strings.Contains(m.message, "Cannot") || strings.Contains(m.message, "Failed") || strings.Contains(m.message, "Error") {
			msgStyle = styles.ErrorStyle
		}
		b.WriteString("\n" + msgStyle.Render(m.message))
	}

	// Help
	helpView := lipgloss.NewStyle().Padding(1, 2).Render(
		"space/enter: toggle • esc/q: back to dashboard",
	)
	b.WriteString("\n" + styles.HelpStyle.Render(helpView))

	return b.String()
}

type BackToDashboardMsg struct{}