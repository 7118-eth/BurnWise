package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"burnwise/internal/models"
	"burnwise/internal/service"
	"burnwise/internal/ui/styles"
)

type categoryListMode int

const (
	categoryListModeView categoryListMode = iota
	categoryListModeEdit
	categoryListModeCreate
	categoryListModeMerge
	categoryListModeConfirmDelete
)

type CategoryListModel struct {
	categoryService *service.CategoryService
	list            list.Model
	categories      []*models.CategoryWithTotal
	mode            categoryListMode
	selectedItem    *categoryItem
	editForm        *CategoryEditModel
	createForm      *CategoryEditModel
	mergeForm       *CategoryMergeModel
	confirmDelete   string
	errorMsg        string
	successMsg      string
}

type categoryItem struct {
	category *models.CategoryWithTotal
}

func (i categoryItem) Title() string {
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

func (i categoryItem) Description() string {
	txCount := "No transactions"
	if i.category.Count > 0 {
		txCount = fmt.Sprintf("%d transactions", i.category.Count)
	}
	
	return fmt.Sprintf("%s · %s", i.category.Type, txCount)
}

func (i categoryItem) FilterValue() string {
	return i.category.Name
}

func NewCategoryListModel(categoryService *service.CategoryService) *CategoryListModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().
		Foreground(lipgloss.Color(styles.PrimaryColor)).
		BorderForeground(lipgloss.Color(styles.PrimaryColor))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Copy().
		Foreground(lipgloss.Color(styles.SecondaryColor)).
		BorderForeground(lipgloss.Color(styles.PrimaryColor))

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "📂 Category Management"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.KeyMap.Quit.SetEnabled(false)

	// Add custom key bindings
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
			key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "merge")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
			key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "history")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		}
	}

	return &CategoryListModel{
		categoryService: categoryService,
		list:            l,
		mode:            categoryListModeView,
	}
}

func (m *CategoryListModel) Init() tea.Cmd {
	return m.loadCategories()
}

func (m *CategoryListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle mode-specific updates
	switch m.mode {
	case categoryListModeEdit:
		if m.editForm != nil {
			newForm, cmd := m.editForm.Update(msg)
			m.editForm = newForm.(*CategoryEditModel)
			
			if m.editForm.completed {
				m.mode = categoryListModeView
				m.successMsg = "Category updated successfully"
				return m, tea.Batch(m.loadCategories(), m.clearMessages())
			} else if m.editForm.cancelled {
				m.mode = categoryListModeView
				m.editForm = nil
			}
			return m, cmd
		}
		
	case categoryListModeCreate:
		if m.createForm != nil {
			newForm, cmd := m.createForm.Update(msg)
			m.createForm = newForm.(*CategoryEditModel)
			
			if m.createForm.completed {
				m.mode = categoryListModeView
				m.successMsg = "Category created successfully"
				return m, tea.Batch(m.loadCategories(), m.clearMessages())
			} else if m.createForm.cancelled {
				m.mode = categoryListModeView
				m.createForm = nil
			}
			return m, cmd
		}
		
	case categoryListModeMerge:
		if m.mergeForm != nil {
			newForm, cmd := m.mergeForm.Update(msg)
			m.mergeForm = newForm.(*CategoryMergeModel)
			
			if m.mergeForm.completed {
				m.mode = categoryListModeView
				m.successMsg = "Categories merged successfully"
				return m, tea.Batch(m.loadCategories(), m.clearMessages())
			} else if m.mergeForm.cancelled {
				m.mode = categoryListModeView
				m.mergeForm = nil
			}
			return m, cmd
		}
		
	case categoryListModeConfirmDelete:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				if m.selectedItem != nil {
					err := m.categoryService.Delete(m.selectedItem.category.ID)
					if err != nil {
						m.errorMsg = err.Error()
					} else {
						m.successMsg = "Category deleted successfully"
					}
				}
				m.mode = categoryListModeView
				m.confirmDelete = ""
				return m, tea.Batch(m.loadCategories(), m.clearMessages())
			case "n", "N", "esc":
				m.mode = categoryListModeView
				m.confirmDelete = ""
			}
		}
		return m, nil
	}

	// Handle main list view
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == categoryListModeView {
			switch msg.String() {
			case "esc", "q":
				// Return to main menu
				return m, nil
			case "n":
				// Create new category
				m.createForm = NewCategoryEditModel(m.categoryService, nil)
				m.mode = categoryListModeCreate
				return m, m.createForm.Init()
			case "e":
				// Edit selected category
				if item, ok := m.list.SelectedItem().(categoryItem); ok {
					if item.category.IsDefault {
						m.errorMsg = "Cannot edit default categories"
						return m, m.clearMessages()
					}
					// Convert CategoryWithTotal to Category for editing
					cat := &models.Category{
						ID:    item.category.ID,
						Name:  item.category.Name,
						Type:  item.category.Type,
						Icon:  item.category.Icon,
						Color: item.category.Color,
					}
					m.editForm = NewCategoryEditModel(m.categoryService, cat)
					m.selectedItem = &item
					m.mode = categoryListModeEdit
					return m, m.editForm.Init()
				}
			case "m":
				// Merge categories
				if item, ok := m.list.SelectedItem().(categoryItem); ok {
					if item.category.IsDefault {
						m.errorMsg = "Cannot merge default categories"
						return m, m.clearMessages()
					}
					if item.category.Count == 0 {
						m.errorMsg = "Category has no transactions to merge"
						return m, m.clearMessages()
					}
					m.mergeForm = NewCategoryMergeModel(m.categoryService, item.category)
					m.selectedItem = &item
					m.mode = categoryListModeMerge
					return m, m.mergeForm.Init()
				}
			case "d":
				// Delete category
				if item, ok := m.list.SelectedItem().(categoryItem); ok {
					if item.category.IsDefault {
						m.errorMsg = "Cannot delete default categories"
						return m, m.clearMessages()
					}
					if item.category.Count > 0 {
						m.errorMsg = fmt.Sprintf("Cannot delete category with %d transactions. Use merge instead.", item.category.Count)
						return m, m.clearMessages()
					}
					m.selectedItem = &item
					m.confirmDelete = fmt.Sprintf("Delete category '%s'? (y/n)", item.category.Name)
					m.mode = categoryListModeConfirmDelete
				}
			case "h":
				// View history (TODO: implement history view)
				if item, ok := m.list.SelectedItem().(categoryItem); ok {
					m.errorMsg = fmt.Sprintf("History view not yet implemented for '%s'", item.category.Name)
					return m, m.clearMessages()
				}
			}
		}
	
	case categoryManagementLoadedMsg:
		m.categories = msg.categories
		items := make([]list.Item, len(m.categories))
		for i, cat := range m.categories {
			items[i] = categoryItem{category: cat}
		}
		m.list.SetItems(items)
		return m, nil
		
	case clearMessagesMsg:
		m.handleClearMessages()
		return m, nil
		
	case tea.WindowSizeMsg:
		h, v := styles.AppStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *CategoryListModel) View() string {
	if m.mode == categoryListModeEdit && m.editForm != nil {
		return m.editForm.View()
	}
	if m.mode == categoryListModeCreate && m.createForm != nil {
		return m.createForm.View()
	}
	if m.mode == categoryListModeMerge && m.mergeForm != nil {
		return m.mergeForm.View()
	}
	
	var content strings.Builder
	content.WriteString(m.list.View())
	
	// Show messages
	if m.errorMsg != "" {
		content.WriteString("\n" + styles.ErrorStyle.Render("❌ "+m.errorMsg))
	}
	if m.successMsg != "" {
		content.WriteString("\n" + styles.SuccessStyle.Render("✅ "+m.successMsg))
	}
	if m.confirmDelete != "" {
		content.WriteString("\n" + styles.WarningStyle.Render("⚠️  "+m.confirmDelete))
	}
	
	return styles.AppStyle.Render(content.String())
}

// Messages
type categoryManagementLoadedMsg struct {
	categories []*models.CategoryWithTotal
}

// Commands
func (m *CategoryListModel) loadCategories() tea.Cmd {
	return func() tea.Msg {
		categories, err := m.categoryService.GetAllWithUsageCount()
		if err != nil {
			return errMsg{err}
		}
		return categoryManagementLoadedMsg{categories: categories}
	}
}

func (m *CategoryListModel) clearMessages() tea.Cmd {
	return tea.Tick(styles.MessageTimeout, func(time.Time) tea.Msg {
		return clearMessagesMsg{}
	})
}

type clearMessagesMsg struct{}

func (m *CategoryListModel) handleClearMessages() {
	m.errorMsg = ""
	m.successMsg = ""
}