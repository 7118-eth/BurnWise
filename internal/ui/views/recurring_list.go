package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"budget-tracker/internal/models"
	"budget-tracker/internal/service"
	"budget-tracker/internal/ui/styles"
)

type recurringListMode int

const (
	recurringListModeView recurringListMode = iota
	recurringListModeEdit
	recurringListModeCreate
	recurringListModeConfirmDelete
	recurringListModeConfirmPause
)

type RecurringListModel struct {
	recurringService *service.RecurringTransactionService
	categoryService  *service.CategoryService
	list             list.Model
	recurringItems   []*models.RecurringTransaction
	mode             recurringListMode
	selectedItem     *recurringItem
	editForm         *RecurringFormModel
	createForm       *RecurringFormModel
	confirmMsg       string
	errorMsg         string
	successMsg       string
}

type recurringItem struct {
	recurring *models.RecurringTransaction
}

func (i recurringItem) Title() string {
	icon := ""
	if i.recurring.Category.Icon != "" {
		icon = i.recurring.Category.Icon + " "
	}
	
	status := ""
	if !i.recurring.IsActive {
		status = " (paused)"
	} else if i.recurring.EndDate != nil && time.Now().After(*i.recurring.EndDate) {
		status = " (ended)"
	}
	
	return fmt.Sprintf("%s%s%s", icon, i.recurring.Description, status)
}

func (i recurringItem) Description() string {
	typeStr := string(i.recurring.Type)
	amountStr := fmt.Sprintf("%s %.2f", i.recurring.Currency, i.recurring.Amount)
	freqStr := i.recurring.GetFrequencyDisplay()
	nextDue := i.recurring.NextDueDate.Format("Jan 2, 2006")
	
	return fmt.Sprintf("%s · %s · %s · Next: %s", typeStr, amountStr, freqStr, nextDue)
}

func (i recurringItem) FilterValue() string {
	return i.recurring.Description
}

func NewRecurringListModel(
	recurringService *service.RecurringTransactionService,
	categoryService *service.CategoryService,
) *RecurringListModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().
		Foreground(lipgloss.Color(styles.PrimaryColor)).
		BorderForeground(lipgloss.Color(styles.PrimaryColor))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Copy().
		Foreground(lipgloss.Color(styles.SecondaryColor)).
		BorderForeground(lipgloss.Color(styles.PrimaryColor))

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "🔄 Recurring Transactions"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.KeyMap.Quit.SetEnabled(false)

	// Add custom key bindings
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
			key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pause/resume")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
			key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view history")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		}
	}

	return &RecurringListModel{
		recurringService: recurringService,
		categoryService:  categoryService,
		list:             l,
		mode:             recurringListModeView,
	}
}

func (m *RecurringListModel) Init() tea.Cmd {
	return m.loadRecurringTransactions()
}

func (m *RecurringListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle mode-specific updates
	switch m.mode {
	case recurringListModeEdit:
		if m.editForm != nil {
			newForm, cmd := m.editForm.Update(msg)
			m.editForm = newForm.(*RecurringFormModel)
			
			if m.editForm.completed {
				m.mode = recurringListModeView
				m.successMsg = "Recurring transaction updated successfully"
				return m, tea.Batch(m.loadRecurringTransactions(), m.clearMessages())
			} else if m.editForm.cancelled {
				m.mode = recurringListModeView
				m.editForm = nil
			}
			return m, cmd
		}
		
	case recurringListModeCreate:
		if m.createForm != nil {
			newForm, cmd := m.createForm.Update(msg)
			m.createForm = newForm.(*RecurringFormModel)
			
			if m.createForm.completed {
				m.mode = recurringListModeView
				m.successMsg = "Recurring transaction created successfully"
				return m, tea.Batch(m.loadRecurringTransactions(), m.clearMessages())
			} else if m.createForm.cancelled {
				m.mode = recurringListModeView
				m.createForm = nil
			}
			return m, cmd
		}
		
	case recurringListModeConfirmDelete:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				if m.selectedItem != nil {
					err := m.recurringService.Delete(m.selectedItem.recurring.ID)
					if err != nil {
						m.errorMsg = err.Error()
					} else {
						m.successMsg = "Recurring transaction deleted successfully"
					}
				}
				m.mode = recurringListModeView
				m.confirmMsg = ""
				return m, tea.Batch(m.loadRecurringTransactions(), m.clearMessages())
			case "n", "N", "esc":
				m.mode = recurringListModeView
				m.confirmMsg = ""
			}
		}
		return m, nil
		
	case recurringListModeConfirmPause:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				if m.selectedItem != nil {
					var err error
					if m.selectedItem.recurring.IsActive {
						err = m.recurringService.Pause(m.selectedItem.recurring.ID)
						if err == nil {
							m.successMsg = "Recurring transaction paused"
						}
					} else {
						err = m.recurringService.Resume(m.selectedItem.recurring.ID)
						if err == nil {
							m.successMsg = "Recurring transaction resumed"
						}
					}
					if err != nil {
						m.errorMsg = err.Error()
					}
				}
				m.mode = recurringListModeView
				m.confirmMsg = ""
				return m, tea.Batch(m.loadRecurringTransactions(), m.clearMessages())
			case "n", "N", "esc":
				m.mode = recurringListModeView
				m.confirmMsg = ""
			}
		}
		return m, nil
	}

	// Handle main list view
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == recurringListModeView {
			switch msg.String() {
			case "esc", "q":
				// Return to main menu
				return m, nil
			case "n":
				// Create new recurring transaction
				m.createForm = NewRecurringFormModel(m.recurringService, m.categoryService, nil)
				m.mode = recurringListModeCreate
				return m, m.createForm.Init()
			case "e":
				// Edit selected recurring transaction
				if item, ok := m.list.SelectedItem().(recurringItem); ok {
					m.editForm = NewRecurringFormModel(m.recurringService, m.categoryService, item.recurring)
					m.selectedItem = &item
					m.mode = recurringListModeEdit
					return m, m.editForm.Init()
				}
			case "p":
				// Pause/resume recurring transaction
				if item, ok := m.list.SelectedItem().(recurringItem); ok {
					m.selectedItem = &item
					if item.recurring.IsActive {
						m.confirmMsg = fmt.Sprintf("Pause recurring transaction '%s'? (y/n)", item.recurring.Description)
					} else {
						m.confirmMsg = fmt.Sprintf("Resume recurring transaction '%s'? (y/n)", item.recurring.Description)
					}
					m.mode = recurringListModeConfirmPause
				}
			case "d":
				// Delete recurring transaction
				if item, ok := m.list.SelectedItem().(recurringItem); ok {
					m.selectedItem = &item
					m.confirmMsg = fmt.Sprintf("Delete recurring transaction '%s'? (y/n)", item.recurring.Description)
					m.mode = recurringListModeConfirmDelete
				}
			case "v":
				// View transaction history (TODO: implement history view)
				if item, ok := m.list.SelectedItem().(recurringItem); ok {
					m.errorMsg = fmt.Sprintf("History view not yet implemented for '%s'", item.recurring.Description)
					return m, m.clearMessages()
				}
			}
		}
	
	case recurringLoadedMsg:
		m.recurringItems = msg.items
		items := make([]list.Item, len(m.recurringItems))
		for i, rt := range m.recurringItems {
			items[i] = recurringItem{recurring: rt}
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

func (m *RecurringListModel) View() string {
	if m.mode == recurringListModeEdit && m.editForm != nil {
		return m.editForm.View()
	}
	if m.mode == recurringListModeCreate && m.createForm != nil {
		return m.createForm.View()
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
	if m.confirmMsg != "" {
		content.WriteString("\n" + styles.WarningStyle.Render("⚠️  "+m.confirmMsg))
	}
	
	return styles.AppStyle.Render(content.String())
}

// Messages
type recurringLoadedMsg struct {
	items []*models.RecurringTransaction
}

// Commands
func (m *RecurringListModel) loadRecurringTransactions() tea.Cmd {
	return func() tea.Msg {
		items, err := m.recurringService.GetAll()
		if err != nil {
			return errMsg{err}
		}
		return recurringLoadedMsg{items: items}
	}
}

func (m *RecurringListModel) clearMessages() tea.Cmd {
	return tea.Tick(styles.MessageTimeout, func(time.Time) tea.Msg {
		return clearMessagesMsg{}
	})
}

func (m *RecurringListModel) handleClearMessages() {
	m.errorMsg = ""
	m.successMsg = ""
}