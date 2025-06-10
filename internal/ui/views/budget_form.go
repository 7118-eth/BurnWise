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

type BudgetForm struct {
	width           int
	height          int
	budgetService   *service.BudgetService
	categoryService *service.CategoryService
	
	editingBudget   *models.Budget
	name            textinput.Model
	amount          textinput.Model
	period          models.BudgetPeriod
	categoryID      uint
	
	categories      []*models.Category
	focusIndex      int
	err             error
}

type BudgetSavedMsg struct{}
type BudgetCancelledMsg struct{}

func NewBudgetForm(budgetService *service.BudgetService, categoryService *service.CategoryService) *BudgetForm {
	name := textinput.New()
	name.Placeholder = "Budget name"
	name.Focus()
	
	amount := textinput.New()
	amount.Placeholder = "0.00"
	
	return &BudgetForm{
		budgetService:   budgetService,
		categoryService: categoryService,
		name:            name,
		amount:          amount,
		period:          models.BudgetPeriodMonthly,
		focusIndex:      0,
	}
}

func (b *BudgetForm) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		b.loadCategories,
	)
}

func (b *BudgetForm) Update(msg tea.Msg) (*BudgetForm, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return b, func() tea.Msg { return BudgetCancelledMsg{} }
		case "tab", "shift+tab":
			b.nextFocus(msg.String() == "shift+tab")
		case "enter":
			if b.focusIndex == 4 { // Save button
				return b, b.save
			} else if b.focusIndex == 5 { // Cancel button
				return b, func() tea.Msg { return BudgetCancelledMsg{} }
			}
		case "p":
			if b.focusIndex == 2 { // Period field
				if b.period == models.BudgetPeriodMonthly {
					b.period = models.BudgetPeriodYearly
				} else {
					b.period = models.BudgetPeriodMonthly
				}
			}
		case "up", "down":
			if b.focusIndex == 3 { // Category field
				b.cycleCategory(msg.String() == "up")
			}
		}
		
	case categoriesLoadedMsg:
		b.categories = msg.categories
		// Filter to expense categories only
		var expenseCategories []*models.Category
		for _, cat := range b.categories {
			if cat.Type == models.TransactionTypeExpense {
				expenseCategories = append(expenseCategories, cat)
			}
		}
		b.categories = expenseCategories
		
		if len(b.categories) > 0 && b.categoryID == 0 {
			b.categoryID = b.categories[0].ID
		}
	}
	
	var cmd tea.Cmd
	b.name, cmd = b.name.Update(msg)
	cmds = append(cmds, cmd)
	
	b.amount, cmd = b.amount.Update(msg)
	cmds = append(cmds, cmd)
	
	return b, tea.Batch(cmds...)
}

func (b *BudgetForm) View() string {
	title := "Create Budget"
	if b.editingBudget != nil {
		title = "Edit Budget"
	}
	title = styles.TitleStyle.Render(title)
	
	nameLabel := styles.FormLabelStyle.Render("Name:")
	nameInput := b.name.View()
	if b.focusIndex == 0 {
		nameInput = styles.FormInputFocusedStyle.Render(nameInput)
	} else {
		nameInput = styles.FormInputStyle.Render(nameInput)
	}
	
	amountLabel := styles.FormLabelStyle.Render("Amount:")
	amountInput := b.amount.View()
	if b.focusIndex == 1 {
		amountInput = styles.FormInputFocusedStyle.Render(amountInput)
	} else {
		amountInput = styles.FormInputStyle.Render(amountInput)
	}
	
	periodLabel := styles.FormLabelStyle.Render("Period:")
	periodValue := string(b.period)
	if b.focusIndex == 2 {
		periodValue = styles.SelectedStyle.Render(periodValue + " (press 'p' to toggle)")
	}
	
	categoryLabel := styles.FormLabelStyle.Render("Category:")
	categoryValue := "Select category"
	if b.categoryID > 0 {
		for _, cat := range b.categories {
			if cat.ID == b.categoryID {
				categoryValue = fmt.Sprintf("%s %s", cat.Icon, cat.Name)
				break
			}
		}
	}
	if b.focusIndex == 3 {
		categoryValue = styles.SelectedStyle.Render(categoryValue + " (↑/↓)")
	}
	
	saveButton := "[Save]"
	cancelButton := "[Cancel]"
	if b.focusIndex == 4 {
		saveButton = styles.ButtonStyle.Render(saveButton)
	} else {
		saveButton = styles.ButtonInactiveStyle.Render(saveButton)
	}
	if b.focusIndex == 5 {
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
		lipgloss.JoinHorizontal(lipgloss.Top, nameLabel, nameInput),
		lipgloss.JoinHorizontal(lipgloss.Top, amountLabel, amountInput),
		lipgloss.JoinHorizontal(lipgloss.Top, periodLabel, periodValue),
		lipgloss.JoinHorizontal(lipgloss.Top, categoryLabel, categoryValue),
		"",
		buttons,
	)
	
	if b.err != nil {
		form += "\n\n" + styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", b.err))
	}
	
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(50).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", form))
	
	return lipgloss.Place(b.width, b.height, lipgloss.Center, lipgloss.Center, box)
}

func (b *BudgetForm) SetSize(width, height int) {
	b.width = width
	b.height = height
}

func (b *BudgetForm) Reset() {
	b.editingBudget = nil
	b.name.SetValue("")
	b.amount.SetValue("")
	b.period = models.BudgetPeriodMonthly
	b.categoryID = 0
	b.focusIndex = 0
	b.err = nil
}

func (b *BudgetForm) SetBudget(budget *models.Budget) {
	b.editingBudget = budget
	b.name.SetValue(budget.Name)
	b.amount.SetValue(fmt.Sprintf("%.2f", budget.Amount))
	b.period = budget.Period
	b.categoryID = budget.CategoryID
	b.focusIndex = 0
	b.err = nil
}

func (b *BudgetForm) nextFocus(reverse bool) {
	if reverse {
		b.focusIndex--
		if b.focusIndex < 0 {
			b.focusIndex = 5
		}
	} else {
		b.focusIndex++
		if b.focusIndex > 5 {
			b.focusIndex = 0
		}
	}
	
	b.name.Blur()
	b.amount.Blur()
	
	switch b.focusIndex {
	case 0:
		b.name.Focus()
	case 1:
		b.amount.Focus()
	}
}

func (b *BudgetForm) cycleCategory(reverse bool) {
	if len(b.categories) == 0 {
		return
	}
	
	currentIdx := 0
	for i, cat := range b.categories {
		if cat.ID == b.categoryID {
			currentIdx = i
			break
		}
	}
	
	if reverse {
		currentIdx--
		if currentIdx < 0 {
			currentIdx = len(b.categories) - 1
		}
	} else {
		currentIdx++
		if currentIdx >= len(b.categories) {
			currentIdx = 0
		}
	}
	
	b.categoryID = b.categories[currentIdx].ID
}

func (b *BudgetForm) save() tea.Msg {
	amount, err := strconv.ParseFloat(b.amount.Value(), 64)
	if err != nil {
		b.err = fmt.Errorf("invalid amount")
		return nil
	}
	
	name := b.name.Value()
	if name == "" {
		name = fmt.Sprintf("%s Budget - %s", b.period, time.Now().Format("January 2006"))
	}
	
	if b.editingBudget != nil {
		// Update existing budget
		b.editingBudget.Name = name
		b.editingBudget.Amount = amount
		b.editingBudget.Period = b.period
		b.editingBudget.CategoryID = b.categoryID
		
		if err := b.budgetService.Update(b.editingBudget); err != nil {
			b.err = err
			return nil
		}
	} else {
		// Create new budget
		budget := &models.Budget{
			Name:       name,
			Amount:     amount,
			Period:     b.period,
			CategoryID: b.categoryID,
			StartDate:  time.Now(),
		}
		
		if err := b.budgetService.Create(budget); err != nil {
			b.err = err
			return nil
		}
	}
	
	return BudgetSavedMsg{}
}

func (b *BudgetForm) loadCategories() tea.Msg {
	categories, _ := b.categoryService.GetAll()
	return categoriesLoadedMsg{categories: categories}
}