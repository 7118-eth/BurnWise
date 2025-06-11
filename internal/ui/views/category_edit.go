package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"budget-tracker/internal/models"
	"budget-tracker/internal/service"
	"budget-tracker/internal/ui/styles"
)

type CategoryEditModel struct {
	categoryService *service.CategoryService
	category        *models.Category
	isEditing       bool
	
	nameInput     textinput.Model
	iconInput     textinput.Model
	colorInput    textinput.Model
	typeSelected  models.TransactionType
	
	focusIndex int
	completed  bool
	cancelled  bool
	errorMsg   string
}

func NewCategoryEditModel(categoryService *service.CategoryService, category *models.Category) *CategoryEditModel {
	isEditing := category != nil
	
	if !isEditing {
		category = &models.Category{
			Type:  models.TransactionTypeExpense, // Default
			Icon:  "📁",
			Color: "#4CAF50",
		}
	}

	// Create input fields
	nameInput := textinput.New()
	nameInput.Placeholder = "Category name"
	nameInput.Focus()
	nameInput.CharLimit = 100
	nameInput.Width = 30
	nameInput.SetValue(category.Name)

	iconInput := textinput.New()
	iconInput.Placeholder = "Icon (emoji)"
	iconInput.CharLimit = 4
	iconInput.Width = 10
	iconInput.SetValue(category.Icon)

	colorInput := textinput.New()
	colorInput.Placeholder = "#FF5722"
	colorInput.CharLimit = 7
	colorInput.Width = 10
	colorInput.SetValue(category.Color)

	return &CategoryEditModel{
		categoryService: categoryService,
		category:        category,
		isEditing:       isEditing,
		nameInput:       nameInput,
		iconInput:       iconInput,
		colorInput:      colorInput,
		typeSelected:    category.Type,
	}
}

func (m *CategoryEditModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *CategoryEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case categoryEditSuccessMsg:
		m.handleSuccess()
		return m, nil
		
	case categoryEditErrorMsg:
		m.handleError(msg.error)
		return m, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.cancelled = true
			return m, nil
			
		case "enter":
			if m.focusIndex == 3 { // Save button focused
				return m, m.save()
			}
			m.nextField()
			
		case "tab":
			m.nextField()
			
		case "shift+tab":
			m.prevField()
			
		case "ctrl+s":
			return m, m.save()
			
		case "1":
			if m.focusIndex == 1 { // Type selection
				m.typeSelected = models.TransactionTypeIncome
			}
		case "2":
			if m.focusIndex == 1 { // Type selection
				m.typeSelected = models.TransactionTypeExpense
			}
		}
	}

	// Update focused input
	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 2:
		m.iconInput, cmd = m.iconInput.Update(msg)
	case 3:
		m.colorInput, cmd = m.colorInput.Update(msg)
	}

	return m, cmd
}

func (m *CategoryEditModel) View() string {
	var b strings.Builder
	
	title := "Create Category"
	if m.isEditing {
		title = "Edit Category"
	}
	
	b.WriteString(styles.TitleStyle.Render(title))
	b.WriteString("\n\n")

	// Name input
	b.WriteString(m.renderField("Name:", m.nameInput.View(), 0))
	b.WriteString("\n")

	// Type selection (only for new categories)
	if !m.isEditing {
		typeStyle := styles.LabelStyle
		if m.focusIndex == 1 {
			typeStyle = styles.FocusedStyle
		}
		
		b.WriteString(typeStyle.Render("Type:"))
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
	} else {
		b.WriteString(styles.LabelStyle.Render("Type: "))
		b.WriteString(string(m.category.Type))
		b.WriteString("\n\n")
	}

	// Icon input
	b.WriteString(m.renderField("Icon:", m.iconInput.View(), 2))
	b.WriteString("\n")

	// Color input
	b.WriteString(m.renderField("Color:", m.colorInput.View(), 3))
	b.WriteString("\n")

	// Action buttons
	if m.focusIndex == 4 {
		b.WriteString(styles.ButtonFocusedStyle.Render("[ Save ]"))
	} else {
		b.WriteString(styles.ButtonStyle.Render("[ Save ]"))
	}
	b.WriteString("  ")
	if m.focusIndex == 5 {
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
	b.WriteString(styles.HelpStyle.Render("Tab: next field • Shift+Tab: prev field • Enter: save • Esc: cancel"))

	return styles.AppStyle.Render(b.String())
}

func (m *CategoryEditModel) renderField(label, input string, index int) string {
	labelStyle := styles.LabelStyle
	if m.focusIndex == index {
		labelStyle = styles.FocusedStyle
	}
	
	return labelStyle.Render(label) + "\n" + input
}

func (m *CategoryEditModel) nextField() {
	maxIndex := 5
	if !m.isEditing {
		maxIndex = 5 // Include type selection
	} else {
		maxIndex = 4 // Skip type selection for editing
	}
	
	m.focusIndex = (m.focusIndex + 1) % (maxIndex + 1)
	m.updateFocus()
}

func (m *CategoryEditModel) prevField() {
	maxIndex := 5
	if !m.isEditing {
		maxIndex = 5
	} else {
		maxIndex = 4
	}
	
	if m.focusIndex == 0 {
		m.focusIndex = maxIndex
	} else {
		m.focusIndex--
	}
	
	if m.isEditing && m.focusIndex == 1 {
		m.focusIndex = 0 // Skip type selection when editing
	}
	
	m.updateFocus()
}

func (m *CategoryEditModel) updateFocus() {
	m.nameInput.Blur()
	m.iconInput.Blur()
	m.colorInput.Blur()

	switch m.focusIndex {
	case 0:
		m.nameInput.Focus()
	case 2:
		m.iconInput.Focus()
	case 3:
		m.colorInput.Focus()
	}
}

func (m *CategoryEditModel) save() tea.Cmd {
	return func() tea.Msg {
		// Validate inputs
		name := strings.TrimSpace(m.nameInput.Value())
		if name == "" {
			return categoryEditErrorMsg{error: fmt.Errorf("category name is required")}
		}

		icon := strings.TrimSpace(m.iconInput.Value())
		if icon == "" {
			icon = "📁"
		}

		color := strings.TrimSpace(m.colorInput.Value())
		if color == "" {
			color = "#4CAF50"
		}

		// Update category
		m.category.Name = name
		m.category.Type = m.typeSelected
		m.category.Icon = icon
		m.category.Color = color

		var err error
		if m.isEditing {
			err = m.categoryService.Update(m.category)
		} else {
			err = m.categoryService.Create(m.category)
		}

		if err != nil {
			return categoryEditErrorMsg{error: err}
		}

		return categoryEditSuccessMsg{}
	}
}

// Messages
type categoryEditSuccessMsg struct{}
type categoryEditErrorMsg struct {
	error error
}

func (m *CategoryEditModel) handleSuccess() {
	m.completed = true
}

func (m *CategoryEditModel) handleError(err error) {
	m.errorMsg = err.Error()
}