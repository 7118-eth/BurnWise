package views

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"budget-tracker/internal/models"
	"budget-tracker/internal/service"
	"budget-tracker/internal/ui/styles"
)

type TransactionForm struct {
	width    int
	height   int
	
	txService       *service.TransactionService
	categoryService *service.CategoryService
	currencyService *service.CurrencyService
	
	editingTx       *models.Transaction
	txType          models.TransactionType
	amount          textinput.Model
	currency        string
	categoryID      uint
	description     textinput.Model
	date            textinput.Model
	
	categories      []*models.Category
	currencies      []string
	
	focusIndex      int
	err             error
}

type TransactionSavedMsg struct{}
type TransactionCancelledMsg struct{}

func NewTransactionForm(
	txService *service.TransactionService,
	categoryService *service.CategoryService,
	currencyService *service.CurrencyService,
) *TransactionForm {
	amount := textinput.New()
	amount.Placeholder = "0.00"
	amount.Focus()
	
	description := textinput.New()
	description.Placeholder = "Description"
	
	date := textinput.New()
	date.Placeholder = "YYYY-MM-DD"
	date.SetValue(time.Now().Format("2006-01-02"))
	
	return &TransactionForm{
		txService:       txService,
		categoryService: categoryService,
		currencyService: currencyService,
		txType:          models.TransactionTypeExpense,
		amount:          amount,
		currency:        "USD",
		description:     description,
		date:            date,
		currencies:      currencyService.GetSupportedCurrencies(),
		focusIndex:      0,
	}
}

func (f *TransactionForm) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		f.loadCategories,
	)
}

func (f *TransactionForm) Update(msg tea.Msg) (*TransactionForm, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return f, func() tea.Msg { return TransactionCancelledMsg{} }
		case "tab", "shift+tab":
			f.nextFocus(msg.String() == "shift+tab")
		case "enter":
			if f.focusIndex == 6 { // Save button
				return f, f.save
			} else if f.focusIndex == 7 { // Cancel button
				return f, func() tea.Msg { return TransactionCancelledMsg{} }
			}
		case "t":
			if f.focusIndex == 0 { // Type field
				if f.txType == models.TransactionTypeExpense {
					f.txType = models.TransactionTypeIncome
				} else {
					f.txType = models.TransactionTypeExpense
				}
				return f, f.loadCategories
			}
		case "c":
			if f.focusIndex == 2 { // Currency field
				currentIdx := 0
				for i, c := range f.currencies {
					if c == f.currency {
						currentIdx = i
						break
					}
				}
				f.currency = f.currencies[(currentIdx+1)%len(f.currencies)]
			}
		case "up", "down":
			if f.focusIndex == 3 { // Category field
				f.cycleCategory(msg.String() == "up")
			}
		}
		
	case categoriesLoadedMsg:
		f.categories = msg.categories
		if len(f.categories) > 0 {
			f.categoryID = f.categories[0].ID
		}
	}
	
	var cmd tea.Cmd
	f.amount, cmd = f.amount.Update(msg)
	cmds = append(cmds, cmd)
	
	f.description, cmd = f.description.Update(msg)
	cmds = append(cmds, cmd)
	
	f.date, cmd = f.date.Update(msg)
	cmds = append(cmds, cmd)
	
	return f, tea.Batch(cmds...)
}

func (f *TransactionForm) View() string {
	title := "Add Transaction"
	if f.editingTx != nil {
		title = "Edit Transaction"
	}
	title = styles.TitleStyle.Render(title)
	
	typeLabel := styles.FormLabelStyle.Render("Type:")
	typeValue := lipgloss.NewStyle().
		Foreground(f.getTypeColor()).
		Render(string(f.txType))
	if f.focusIndex == 0 {
		typeValue = styles.SelectedStyle.Render(typeValue + " (press 't' to toggle)")
	}
	
	amountLabel := styles.FormLabelStyle.Render("Amount:")
	amountInput := f.amount.View()
	if f.focusIndex == 1 {
		amountInput = styles.FormInputFocusedStyle.Render(amountInput)
	} else {
		amountInput = styles.FormInputStyle.Render(amountInput)
	}
	
	currencyLabel := styles.FormLabelStyle.Render("Currency:")
	currencyValue := f.currency
	if f.focusIndex == 2 {
		currencyValue = styles.SelectedStyle.Render(currencyValue + " (press 'c' to change)")
	}
	
	categoryLabel := styles.FormLabelStyle.Render("Category:")
	categoryValue := "Select category"
	if f.categoryID > 0 {
		for _, cat := range f.categories {
			if cat.ID == f.categoryID {
				categoryValue = fmt.Sprintf("%s %s", cat.Icon, cat.Name)
				break
			}
		}
	}
	if f.focusIndex == 3 {
		categoryValue = styles.SelectedStyle.Render(categoryValue + " (↑/↓)")
	}
	
	descLabel := styles.FormLabelStyle.Render("Description:")
	descInput := f.description.View()
	if f.focusIndex == 4 {
		descInput = styles.FormInputFocusedStyle.Render(descInput)
	} else {
		descInput = styles.FormInputStyle.Render(descInput)
	}
	
	dateLabel := styles.FormLabelStyle.Render("Date:")
	dateInput := f.date.View()
	if f.focusIndex == 5 {
		dateInput = styles.FormInputFocusedStyle.Render(dateInput)
	} else {
		dateInput = styles.FormInputStyle.Render(dateInput)
	}
	
	saveButton := "[Save]"
	cancelButton := "[Cancel]"
	if f.focusIndex == 6 {
		saveButton = styles.ButtonStyle.Render(saveButton)
	} else {
		saveButton = styles.ButtonInactiveStyle.Render(saveButton)
	}
	if f.focusIndex == 7 {
		cancelButton = styles.ButtonStyle.Render(cancelButton)
	} else {
		cancelButton = styles.ButtonInactiveStyle.Render(cancelButton)
	}
	
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		saveButton,
		"  ",
		cancelButton,
	)
	
	form := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, typeLabel, typeValue),
		lipgloss.JoinHorizontal(lipgloss.Top, amountLabel, amountInput),
		lipgloss.JoinHorizontal(lipgloss.Top, currencyLabel, currencyValue),
		lipgloss.JoinHorizontal(lipgloss.Top, categoryLabel, categoryValue),
		lipgloss.JoinHorizontal(lipgloss.Top, descLabel, descInput),
		lipgloss.JoinHorizontal(lipgloss.Top, dateLabel, dateInput),
		"",
		buttons,
	)
	
	if f.err != nil {
		form += "\n\n" + styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", f.err))
	}
	
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(50).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", form))
	
	return lipgloss.Place(f.width, f.height, lipgloss.Center, lipgloss.Center, box)
}

func (f *TransactionForm) SetSize(width, height int) {
	f.width = width
	f.height = height
}

func (f *TransactionForm) Reset() {
	f.editingTx = nil
	f.txType = models.TransactionTypeExpense
	f.amount.SetValue("")
	f.currency = "USD"
	f.categoryID = 0
	f.description.SetValue("")
	f.date.SetValue(time.Now().Format("2006-01-02"))
	f.focusIndex = 0
	f.err = nil
}

func (f *TransactionForm) SetTransaction(tx *models.Transaction) {
	f.editingTx = tx
	f.txType = tx.Type
	f.amount.SetValue(fmt.Sprintf("%.2f", tx.Amount))
	f.currency = tx.Currency
	f.categoryID = tx.CategoryID
	f.description.SetValue(tx.Description)
	f.date.SetValue(tx.Date.Format("2006-01-02"))
	f.focusIndex = 0
	f.err = nil
}

func (f *TransactionForm) nextFocus(reverse bool) {
	if reverse {
		f.focusIndex--
		if f.focusIndex < 0 {
			f.focusIndex = 7
		}
	} else {
		f.focusIndex++
		if f.focusIndex > 7 {
			f.focusIndex = 0
		}
	}
	
	f.amount.Blur()
	f.description.Blur()
	f.date.Blur()
	
	switch f.focusIndex {
	case 1:
		f.amount.Focus()
	case 4:
		f.description.Focus()
	case 5:
		f.date.Focus()
	}
}

func (f *TransactionForm) cycleCategory(reverse bool) {
	if len(f.categories) == 0 {
		return
	}
	
	currentIdx := 0
	for i, cat := range f.categories {
		if cat.ID == f.categoryID {
			currentIdx = i
			break
		}
	}
	
	if reverse {
		currentIdx--
		if currentIdx < 0 {
			currentIdx = len(f.categories) - 1
		}
	} else {
		currentIdx++
		if currentIdx >= len(f.categories) {
			currentIdx = 0
		}
	}
	
	f.categoryID = f.categories[currentIdx].ID
}

func (f *TransactionForm) getTypeColor() lipgloss.Color {
	if f.txType == models.TransactionTypeIncome {
		return styles.Income
	}
	return styles.Expense
}

func (f *TransactionForm) save() tea.Msg {
	amount, err := strconv.ParseFloat(f.amount.Value(), 64)
	if err != nil {
		f.err = fmt.Errorf("invalid amount")
		return nil
	}
	
	date, err := time.Parse("2006-01-02", f.date.Value())
	if err != nil {
		f.err = fmt.Errorf("invalid date format")
		return nil
	}
	
	if f.editingTx != nil {
		// Update existing transaction
		f.editingTx.Type = f.txType
		f.editingTx.Amount = amount
		f.editingTx.Currency = f.currency
		f.editingTx.CategoryID = f.categoryID
		f.editingTx.Description = f.description.Value()
		f.editingTx.Date = date
		
		if err := f.txService.Update(f.editingTx); err != nil {
			f.err = err
			return nil
		}
	} else {
		// Create new transaction
		tx := &models.Transaction{
			Type:        f.txType,
			Amount:      amount,
			Currency:    f.currency,
			CategoryID:  f.categoryID,
			Description: f.description.Value(),
			Date:        date,
		}
		
		if err := f.txService.Create(tx); err != nil {
			f.err = err
			return nil
		}
	}
	
	return TransactionSavedMsg{}
}

func (f *TransactionForm) loadCategories() tea.Msg {
	categories, _ := f.categoryService.GetByType(f.txType)
	return categoriesLoadedMsg{categories: categories}
}

type categoriesLoadedMsg struct {
	categories []*models.Category
}