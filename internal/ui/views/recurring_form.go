package views

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"burnwise/internal/models"
	"burnwise/internal/service"
	"burnwise/internal/ui/styles"
)

type RecurringFormModel struct {
	recurringService *service.RecurringTransactionService
	categoryService  *service.CategoryService
	recurring        *models.RecurringTransaction
	isEditing        bool
	
	// Form fields
	descriptionInput   textinput.Model
	amountInput        textinput.Model
	frequencyValueInput textinput.Model
	startDateInput     textinput.Model
	endDateInput       textinput.Model
	
	// Selections
	typeSelected       models.TransactionType
	categorySelected   uint
	currencySelected   string
	frequencySelected  models.RecurrenceFrequency
	categories         []*models.Category
	currencies         []string
	
	focusIndex int
	completed  bool
	cancelled  bool
	errorMsg   string
}

func (m *RecurringFormModel) IsCompleted() bool {
	return m.completed
}

func (m *RecurringFormModel) IsCancelled() bool {
	return m.cancelled
}

func NewRecurringFormModel(
	recurringService *service.RecurringTransactionService,
	categoryService *service.CategoryService,
	recurring *models.RecurringTransaction,
) *RecurringFormModel {
	isEditing := recurring != nil
	
	if !isEditing {
		recurring = &models.RecurringTransaction{
			Type:           models.TransactionTypeExpense,
			Currency:       "USD",
			Frequency:      models.FrequencyMonthly,
			FrequencyValue: 1,
			StartDate:      time.Now(),
			IsActive:       true,
		}
	}

	// Create input fields
	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "Rent, AI Tools, Cloud Services, etc."
	descriptionInput.Focus()
	descriptionInput.CharLimit = 255
	descriptionInput.Width = 50
	descriptionInput.SetValue(recurring.Description)

	amountInput := textinput.New()
	amountInput.Placeholder = "0.00"
	amountInput.CharLimit = 15
	amountInput.Width = 20
	if recurring.Amount > 0 {
		amountInput.SetValue(fmt.Sprintf("%.2f", recurring.Amount))
	}

	frequencyValueInput := textinput.New()
	frequencyValueInput.Placeholder = "1"
	frequencyValueInput.CharLimit = 3
	frequencyValueInput.Width = 10
	frequencyValueInput.SetValue(strconv.Itoa(recurring.FrequencyValue))

	startDateInput := textinput.New()
	startDateInput.Placeholder = "YYYY-MM-DD"
	startDateInput.CharLimit = 10
	startDateInput.Width = 15
	startDateInput.SetValue(recurring.StartDate.Format("2006-01-02"))

	endDateInput := textinput.New()
	endDateInput.Placeholder = "YYYY-MM-DD (optional)"
	endDateInput.CharLimit = 10
	endDateInput.Width = 15
	if recurring.EndDate != nil {
		endDateInput.SetValue(recurring.EndDate.Format("2006-01-02"))
	}

	// Default currencies - in real app, this would come from settings
	currencies := []string{"USD", "EUR", "AED"}

	return &RecurringFormModel{
		recurringService:    recurringService,
		categoryService:     categoryService,
		recurring:           recurring,
		isEditing:           isEditing,
		descriptionInput:    descriptionInput,
		amountInput:         amountInput,
		frequencyValueInput: frequencyValueInput,
		startDateInput:      startDateInput,
		endDateInput:        endDateInput,
		typeSelected:        recurring.Type,
		categorySelected:    recurring.CategoryID,
		currencySelected:    recurring.Currency,
		frequencySelected:   recurring.Frequency,
		currencies:          currencies,
	}
}

func (m *RecurringFormModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.loadCategories(),
	)
}

func (m *RecurringFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case categoriesLoadedForRecurringMsg:
		m.categories = msg.categories
		// Set default category if not editing
		if !m.isEditing && len(m.categories) > 0 && m.categorySelected == 0 {
			for _, cat := range m.categories {
				if cat.Type == m.typeSelected {
					m.categorySelected = cat.ID
					break
				}
			}
		}
		return m, nil
		
	case recurringFormSuccessMsg:
		m.completed = true
		return m, nil
		
	case recurringFormErrorMsg:
		m.errorMsg = msg.error.Error()
		return m, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.cancelled = true
			return m, nil
			
		case "ctrl+s":
			return m, m.save()
			
		case "tab":
			m.nextField()
			
		case "shift+tab":
			m.prevField()
			
		case "enter":
			if m.focusIndex == 10 { // Save button
				return m, m.save()
			} else if m.focusIndex == 11 { // Cancel button
				m.cancelled = true
				return m, nil
			}
			m.nextField()
			
		// Type selection
		case "1":
			if m.focusIndex == 1 {
				m.typeSelected = models.TransactionTypeIncome
				m.updateCategoriesForType()
			}
		case "2":
			if m.focusIndex == 1 {
				m.typeSelected = models.TransactionTypeExpense
				m.updateCategoriesForType()
			}
			
		// Frequency selection
		case "d":
			if m.focusIndex == 5 {
				m.frequencySelected = models.FrequencyDaily
			}
		case "w":
			if m.focusIndex == 5 {
				m.frequencySelected = models.FrequencyWeekly
			}
		case "m":
			if m.focusIndex == 5 {
				m.frequencySelected = models.FrequencyMonthly
			}
		case "y":
			if m.focusIndex == 5 {
				m.frequencySelected = models.FrequencyYearly
			}
			
		// Category navigation
		case "j", "down":
			if m.focusIndex == 3 {
				m.nextCategory()
			}
		case "k", "up":
			if m.focusIndex == 3 {
				m.prevCategory()
			}
			
		// Currency navigation
		case "left":
			if m.focusIndex == 4 {
				m.prevCurrency()
			}
		case "right":
			if m.focusIndex == 4 {
				m.nextCurrency()
			}
		}
	}

	// Update focused input
	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.descriptionInput, cmd = m.descriptionInput.Update(msg)
	case 2:
		m.amountInput, cmd = m.amountInput.Update(msg)
	case 6:
		m.frequencyValueInput, cmd = m.frequencyValueInput.Update(msg)
	case 7:
		m.startDateInput, cmd = m.startDateInput.Update(msg)
	case 8:
		m.endDateInput, cmd = m.endDateInput.Update(msg)
	}

	return m, cmd
}

func (m *RecurringFormModel) View() string {
	var b strings.Builder
	
	title := "Create Recurring Transaction"
	if m.isEditing {
		title = "Edit Recurring Transaction"
	}
	
	b.WriteString(styles.TitleStyle.Render(title))
	b.WriteString("\n\n")

	// Description
	b.WriteString(m.renderField("Description:", m.descriptionInput.View(), 0))
	b.WriteString("\n")

	// Type selection
	b.WriteString(m.renderField("Type:", "", 1))
	b.WriteString("\n")
	
	incomeStyle := styles.OptionStyle
	expenseStyle := styles.OptionStyle
	
	if m.typeSelected == models.TransactionTypeIncome {
		incomeStyle = styles.SelectedStyle
	} else {
		expenseStyle = styles.SelectedStyle
	}
	
	b.WriteString("  " + incomeStyle.Render("[1] Income") + "  " + expenseStyle.Render("[2] Expense"))
	b.WriteString("\n\n")

	// Amount
	b.WriteString(m.renderField("Amount:", m.amountInput.View(), 2))
	b.WriteString("\n")

	// Category selection
	categoryDisplay := "No category selected"
	if m.categorySelected > 0 {
		for _, cat := range m.categories {
			if cat.ID == m.categorySelected {
				icon := cat.Icon
				if icon == "" {
					icon = "📁"
				}
				categoryDisplay = fmt.Sprintf("%s %s", icon, cat.Name)
				break
			}
		}
	}
	
	b.WriteString(m.renderField("Category:", categoryDisplay, 3))
	if m.focusIndex == 3 {
		b.WriteString("\n  " + styles.HelpStyle.Render("↑/↓ to navigate"))
	}
	b.WriteString("\n")

	// Currency
	b.WriteString(m.renderField("Currency:", m.currencySelected, 4))
	if m.focusIndex == 4 {
		b.WriteString("\n  " + styles.HelpStyle.Render("←/→ to change"))
	}
	b.WriteString("\n")

	// Frequency
	b.WriteString(m.renderField("Frequency:", "", 5))
	b.WriteString("\n")
	
	freqOptions := []struct {
		key  string
		freq models.RecurrenceFrequency
		name string
	}{
		{"d", models.FrequencyDaily, "Daily"},
		{"w", models.FrequencyWeekly, "Weekly"},
		{"m", models.FrequencyMonthly, "Monthly"},
		{"y", models.FrequencyYearly, "Yearly"},
	}
	
	for i, opt := range freqOptions {
		if i > 0 {
			b.WriteString("  ")
		}
		style := styles.OptionStyle
		if m.frequencySelected == opt.freq {
			style = styles.SelectedStyle
		}
		b.WriteString("  " + style.Render(fmt.Sprintf("[%s] %s", opt.key, opt.name)))
	}
	b.WriteString("\n\n")

	// Frequency value
	b.WriteString(m.renderField("Every:", m.frequencyValueInput.View()+" "+m.getFrequencyUnit(), 6))
	b.WriteString("\n")

	// Start date
	b.WriteString(m.renderField("Start Date:", m.startDateInput.View(), 7))
	b.WriteString("\n")

	// End date
	b.WriteString(m.renderField("End Date:", m.endDateInput.View(), 8))
	b.WriteString("\n")

	// Action buttons
	b.WriteString("\n")
	if m.focusIndex == 10 {
		b.WriteString(styles.ButtonFocusedStyle.Render("[ Save ]"))
	} else {
		b.WriteString(styles.ButtonStyle.Render("[ Save ]"))
	}
	b.WriteString("  ")
	if m.focusIndex == 11 {
		b.WriteString(styles.ButtonFocusedStyle.Render("[ Cancel ]"))
	} else {
		b.WriteString(styles.ButtonStyle.Render("[ Cancel ]"))
	}

	// Error message
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		b.WriteString(styles.ErrorStyle.Render("❌ " + m.errorMsg))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("Tab: next field • Shift+Tab: prev field • Ctrl+S: save • Esc: cancel"))

	return styles.AppStyle.Render(b.String())
}

func (m *RecurringFormModel) renderField(label, value string, index int) string {
	labelStyle := styles.LabelStyle
	if m.focusIndex == index {
		labelStyle = styles.FocusedStyle
	}
	
	result := labelStyle.Render(label)
	if value != "" {
		result += "\n" + value
	}
	return result
}

func (m *RecurringFormModel) nextField() {
	m.focusIndex = (m.focusIndex + 1) % 12
	m.updateFocus()
}

func (m *RecurringFormModel) prevField() {
	if m.focusIndex == 0 {
		m.focusIndex = 11
	} else {
		m.focusIndex--
	}
	m.updateFocus()
}

func (m *RecurringFormModel) updateFocus() {
	m.descriptionInput.Blur()
	m.amountInput.Blur()
	m.frequencyValueInput.Blur()
	m.startDateInput.Blur()
	m.endDateInput.Blur()

	switch m.focusIndex {
	case 0:
		m.descriptionInput.Focus()
	case 2:
		m.amountInput.Focus()
	case 6:
		m.frequencyValueInput.Focus()
	case 7:
		m.startDateInput.Focus()
	case 8:
		m.endDateInput.Focus()
	}
}

func (m *RecurringFormModel) updateCategoriesForType() {
	// Reset category selection when type changes
	m.categorySelected = 0
	for _, cat := range m.categories {
		if cat.Type == m.typeSelected {
			m.categorySelected = cat.ID
			break
		}
	}
}

func (m *RecurringFormModel) nextCategory() {
	availableCategories := m.getAvailableCategories()
	if len(availableCategories) == 0 {
		return
	}

	currentIndex := -1
	for i, cat := range availableCategories {
		if cat.ID == m.categorySelected {
			currentIndex = i
			break
		}
	}

	nextIndex := (currentIndex + 1) % len(availableCategories)
	m.categorySelected = availableCategories[nextIndex].ID
}

func (m *RecurringFormModel) prevCategory() {
	availableCategories := m.getAvailableCategories()
	if len(availableCategories) == 0 {
		return
	}

	currentIndex := -1
	for i, cat := range availableCategories {
		if cat.ID == m.categorySelected {
			currentIndex = i
			break
		}
	}

	prevIndex := currentIndex - 1
	if prevIndex < 0 {
		prevIndex = len(availableCategories) - 1
	}
	m.categorySelected = availableCategories[prevIndex].ID
}

func (m *RecurringFormModel) getAvailableCategories() []*models.Category {
	var available []*models.Category
	for _, cat := range m.categories {
		if cat.Type == m.typeSelected {
			available = append(available, cat)
		}
	}
	return available
}

func (m *RecurringFormModel) nextCurrency() {
	for i, curr := range m.currencies {
		if curr == m.currencySelected {
			m.currencySelected = m.currencies[(i+1)%len(m.currencies)]
			break
		}
	}
}

func (m *RecurringFormModel) prevCurrency() {
	for i, curr := range m.currencies {
		if curr == m.currencySelected {
			idx := i - 1
			if idx < 0 {
				idx = len(m.currencies) - 1
			}
			m.currencySelected = m.currencies[idx]
			break
		}
	}
}

func (m *RecurringFormModel) getFrequencyUnit() string {
	switch m.frequencySelected {
	case models.FrequencyDaily:
		return "day(s)"
	case models.FrequencyWeekly:
		return "week(s)"
	case models.FrequencyMonthly:
		return "month(s)"
	case models.FrequencyYearly:
		return "year(s)"
	default:
		return ""
	}
}

func (m *RecurringFormModel) save() tea.Cmd {
	return func() tea.Msg {
		// Validate inputs
		description := strings.TrimSpace(m.descriptionInput.Value())
		if description == "" {
			return recurringFormErrorMsg{error: fmt.Errorf("description is required")}
		}

		amountStr := strings.TrimSpace(m.amountInput.Value())
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || amount <= 0 {
			return recurringFormErrorMsg{error: fmt.Errorf("valid amount is required")}
		}

		if m.categorySelected == 0 {
			return recurringFormErrorMsg{error: fmt.Errorf("category is required")}
		}

		freqValueStr := strings.TrimSpace(m.frequencyValueInput.Value())
		freqValue, err := strconv.Atoi(freqValueStr)
		if err != nil || freqValue < 1 {
			return recurringFormErrorMsg{error: fmt.Errorf("frequency value must be at least 1")}
		}

		// Parse dates
		startDateStr := strings.TrimSpace(m.startDateInput.Value())
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return recurringFormErrorMsg{error: fmt.Errorf("invalid start date format (use YYYY-MM-DD)")}
		}

		var endDate *time.Time
		endDateStr := strings.TrimSpace(m.endDateInput.Value())
		if endDateStr != "" {
			ed, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				return recurringFormErrorMsg{error: fmt.Errorf("invalid end date format (use YYYY-MM-DD)")}
			}
			endDate = &ed
		}

		// Update recurring transaction
		m.recurring.Type = m.typeSelected
		m.recurring.Amount = amount
		m.recurring.Currency = m.currencySelected
		m.recurring.CategoryID = m.categorySelected
		m.recurring.Description = description
		m.recurring.Frequency = m.frequencySelected
		m.recurring.FrequencyValue = freqValue
		m.recurring.StartDate = startDate
		m.recurring.EndDate = endDate

		var err2 error
		if m.isEditing {
			err2 = m.recurringService.Update(m.recurring)
		} else {
			err2 = m.recurringService.Create(m.recurring)
		}

		if err2 != nil {
			return recurringFormErrorMsg{error: err2}
		}

		return recurringFormSuccessMsg{}
	}
}

// Commands
func (m *RecurringFormModel) loadCategories() tea.Cmd {
	return func() tea.Msg {
		categories, err := m.categoryService.GetAll()
		if err != nil {
			return recurringFormErrorMsg{error: err}
		}
		return categoriesLoadedForRecurringMsg{categories: categories}
	}
}

// Messages
type recurringFormSuccessMsg struct{}
type recurringFormErrorMsg struct {
	error error
}
type categoriesLoadedForRecurringMsg struct {
	categories []*models.Category
}