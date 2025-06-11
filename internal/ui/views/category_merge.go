package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"budget-tracker/internal/models"
	"budget-tracker/internal/service"
	"budget-tracker/internal/ui/styles"
)

type CategoryMergeModel struct {
	categoryService *service.CategoryService
	sourceCategory  *models.CategoryWithTotal
	targetList      list.Model
	targetCategories []*models.CategoryWithTotal
	completed       bool
	cancelled       bool
	errorMsg        string
	confirmMerge    bool
	selectedTarget  *models.CategoryWithTotal
}

type mergeTargetItem struct {
	category *models.CategoryWithTotal
}

func (i mergeTargetItem) Title() string {
	icon := i.category.Icon
	if icon == "" {
		icon = "📁"
	}
	
	status := ""
	if i.category.IsDefault {
		status = " (default)"
	}
	
	return fmt.Sprintf("%s %s%s", icon, i.category.Name, status)
}

func (i mergeTargetItem) Description() string {
	txCount := "No transactions"
	if i.category.Count > 0 {
		txCount = fmt.Sprintf("%d transactions", i.category.Count)
	}
	
	return fmt.Sprintf("%s · %s", i.category.Type, txCount)
}

func (i mergeTargetItem) FilterValue() string {
	return i.category.Name
}

func NewCategoryMergeModel(categoryService *service.CategoryService, sourceCategory *models.CategoryWithTotal) *CategoryMergeModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().
		Foreground(lipgloss.Color(styles.PrimaryColor)).
		BorderForeground(lipgloss.Color(styles.PrimaryColor))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Copy().
		Foreground(lipgloss.Color(styles.SecondaryColor)).
		BorderForeground(lipgloss.Color(styles.PrimaryColor))

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = fmt.Sprintf("Select Target Category for '%s'", sourceCategory.Name)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.KeyMap.Quit.SetEnabled(false)

	return &CategoryMergeModel{
		categoryService: categoryService,
		sourceCategory:  sourceCategory,
		targetList:      l,
	}
}

func (m *CategoryMergeModel) Init() tea.Cmd {
	return m.loadTargetCategories()
}

func (m *CategoryMergeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.confirmMerge {
			switch msg.String() {
			case "y", "Y":
				return m, m.performMerge()
			case "n", "N", "esc":
				m.confirmMerge = false
				m.selectedTarget = nil
			}
			return m, nil
		}
		
		switch msg.String() {
		case "esc":
			m.cancelled = true
			return m, nil
			
		case "enter":
			if item, ok := m.targetList.SelectedItem().(mergeTargetItem); ok {
				m.selectedTarget = item.category
				m.confirmMerge = true
			}
		}
		
	case targetCategoriesLoadedMsg:
		m.targetCategories = msg.categories
		items := make([]list.Item, 0)
		
		// Only include categories of the same type, excluding the source category
		for _, cat := range m.targetCategories {
			if cat.Type == m.sourceCategory.Type && cat.ID != m.sourceCategory.ID {
				items = append(items, mergeTargetItem{category: cat})
			}
		}
		
		m.targetList.SetItems(items)
		return m, nil
		
	case categoryMergeSuccessMsg:
		m.completed = true
		return m, nil
		
	case categoryMergeErrorMsg:
		m.errorMsg = msg.error.Error()
		m.confirmMerge = false
		m.selectedTarget = nil
		return m, nil
		
	case tea.WindowSizeMsg:
		h, v := styles.AppStyle.GetFrameSize()
		m.targetList.SetSize(msg.Width-h, msg.Height-v-8)
	}

	var cmd tea.Cmd
	m.targetList, cmd = m.targetList.Update(msg)
	return m, cmd
}

func (m *CategoryMergeModel) View() string {
	var b strings.Builder
	
	b.WriteString(styles.TitleStyle.Render("Merge Categories"))
	b.WriteString("\n\n")

	// Source category info
	sourceIcon := m.sourceCategory.Icon
	if sourceIcon == "" {
		sourceIcon = "📁"
	}
	
	b.WriteString(styles.LabelStyle.Render("Source Category:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s %s (%s, %d transactions)", 
		sourceIcon, m.sourceCategory.Name, m.sourceCategory.Type, m.sourceCategory.Count))
	b.WriteString("\n\n")

	if m.confirmMerge && m.selectedTarget != nil {
		// Confirmation dialog
		targetIcon := m.selectedTarget.Icon
		if targetIcon == "" {
			targetIcon = "📁"
		}
		
		b.WriteString(styles.WarningStyle.Render("⚠️  CONFIRM MERGE"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("Merge '%s' into '%s'?", m.sourceCategory.Name, m.selectedTarget.Name))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("This will move %d transactions from '%s' to '%s'", 
			m.sourceCategory.Count, m.sourceCategory.Name, m.selectedTarget.Name))
		b.WriteString("\n")
		b.WriteString(styles.ErrorStyle.Render("This action cannot be undone!"))
		b.WriteString("\n\n")
		b.WriteString("Continue? (y/n)")
		
	} else {
		// Target selection
		b.WriteString(styles.LabelStyle.Render("Select target category:"))
		b.WriteString("\n")
		b.WriteString(m.targetList.View())
		
		if len(m.targetCategories) == 0 {
			b.WriteString("\n")
			b.WriteString(styles.WarningStyle.Render("No compatible categories found for merging."))
		}
	}

	// Error message
	if m.errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(styles.ErrorStyle.Render("❌ " + m.errorMsg))
	}

	// Help
	if !m.confirmMerge {
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Enter: select • Esc: cancel"))
	}

	return styles.AppStyle.Render(b.String())
}

// Commands
func (m *CategoryMergeModel) loadTargetCategories() tea.Cmd {
	return func() tea.Msg {
		categories, err := m.categoryService.GetAllWithUsageCount()
		if err != nil {
			return categoryMergeErrorMsg{error: err}
		}
		return targetCategoriesLoadedMsg{categories: categories}
	}
}

func (m *CategoryMergeModel) performMerge() tea.Cmd {
	return func() tea.Msg {
		if m.selectedTarget == nil {
			return categoryMergeErrorMsg{error: fmt.Errorf("no target category selected")}
		}
		
		err := m.categoryService.MergeCategories(m.sourceCategory.ID, m.selectedTarget.ID)
		if err != nil {
			return categoryMergeErrorMsg{error: err}
		}
		
		return categoryMergeSuccessMsg{}
	}
}

// Messages
type targetCategoriesLoadedMsg struct {
	categories []*models.CategoryWithTotal
}

type categoryMergeSuccessMsg struct{}

type categoryMergeErrorMsg struct {
	error error
}