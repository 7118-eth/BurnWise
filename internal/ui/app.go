package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"burnwise/internal/service"
	"burnwise/internal/ui/views"
)

type view int

const (
	viewDashboard view = iota
	viewTransactions
	viewTransactionForm
	viewBudgets
	viewBudgetForm
	viewReports
	viewCategories
	viewRecurring
	viewRecurringForm
	viewCurrencySettings
)

type App struct {
	currentView     view
	width           int
	height          int
	
	txService              *service.TransactionService
	categoryService        *service.CategoryService
	budgetService          *service.BudgetService
	currencyService        *service.CurrencyService
	settingsService        *service.SettingsService
	recurringService       *service.RecurringTransactionService
	
	dashboard        *views.Dashboard
	transactionList  *views.TransactionList
	transactionForm  *views.TransactionForm
	budgetList       *views.BudgetList
	budgetForm       *views.BudgetForm
	reports          *views.Reports
	categoryList     *views.CategoryListModel
	recurringList    *views.RecurringListModel
	recurringForm    *views.RecurringFormModel
	currencySettings *views.CurrencySettings
	
	err             error
}

func NewApp(
	txService *service.TransactionService,
	categoryService *service.CategoryService,
	budgetService *service.BudgetService,
	currencyService *service.CurrencyService,
	settingsService *service.SettingsService,
	recurringService *service.RecurringTransactionService,
) *App {
	return &App{
		currentView:      viewDashboard,
		txService:        txService,
		categoryService:  categoryService,
		budgetService:    budgetService,
		currencyService:  currencyService,
		settingsService:  settingsService,
		recurringService: recurringService,
	}
}

func (a *App) Init() tea.Cmd {
	a.dashboard = views.NewDashboard(a.txService, a.budgetService)
	a.transactionList = views.NewTransactionList(a.txService, a.categoryService)
	a.transactionForm = views.NewTransactionForm(a.txService, a.categoryService, a.currencyService)
	a.budgetList = views.NewBudgetList(a.budgetService, a.categoryService)
	a.budgetForm = views.NewBudgetForm(a.budgetService, a.categoryService)
	a.reports = views.NewReports(a.txService, a.categoryService, a.budgetService)
	a.categoryList = views.NewCategoryListModel(a.categoryService)
	a.recurringList = views.NewRecurringListModel(a.recurringService, a.categoryService)
	a.currencySettings = views.NewCurrencySettings(a.settingsService, a.currencyService, a.txService)
	
	return tea.Batch(
		a.dashboard.Init(),
		tea.EnterAltScreen,
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateViewSizes()

	case tea.KeyMsg:
		if a.currentView == viewDashboard || a.currentView == viewTransactions || 
		   a.currentView == viewBudgets || a.currentView == viewReports || 
		   a.currentView == viewCategories || a.currentView == viewRecurring {
			switch msg.String() {
			case "q", "ctrl+c":
				return a, tea.Quit
			case "n":
				if a.currentView == viewDashboard || a.currentView == viewTransactions {
					a.currentView = viewTransactionForm
					a.transactionForm.Reset()
					return a, a.transactionForm.Init()
				} else if a.currentView == viewBudgets {
					a.currentView = viewBudgetForm
					a.budgetForm.Reset()
					return a, a.budgetForm.Init()
				} else if a.currentView == viewRecurring {
					a.currentView = viewRecurringForm
					a.recurringForm = views.NewRecurringFormModel(a.recurringService, a.categoryService, nil)
					return a, a.recurringForm.Init()
				}
			case "t":
				a.currentView = viewTransactions
				return a, a.transactionList.Init()
			case "b":
				a.currentView = viewBudgets
				return a, a.budgetList.Init()
			case "r":
				a.currentView = viewReports
				return a, a.reports.Init()
			case "c":
				a.currentView = viewCategories
				return a, a.categoryList.Init()
			case "u":
				a.currentView = viewCurrencySettings
				return a, a.currencySettings.Init()
			case "s":
				a.currentView = viewRecurring
				return a, a.recurringList.Init()
			case "esc":
				a.currentView = viewDashboard
				return a, a.dashboard.Init()
			}
		}

	case views.TransactionSavedMsg:
		a.currentView = viewDashboard
		return a, a.dashboard.Init()
		
	case views.TransactionCancelledMsg:
		if a.transactionList.HasTransactions() {
			a.currentView = viewTransactions
			return a, a.transactionList.Init()
		} else {
			a.currentView = viewDashboard
			return a, a.dashboard.Init()
		}
		
	case views.TransactionEditMsg:
		a.currentView = viewTransactionForm
		a.transactionForm.SetTransaction(msg.Transaction)
		return a, a.transactionForm.Init()
		
	case views.BudgetSavedMsg:
		a.currentView = viewBudgets
		return a, a.budgetList.Init()
		
	case views.BudgetCancelledMsg:
		a.currentView = viewBudgets
		return a, a.budgetList.Init()
		
	case views.BudgetEditMsg:
		a.currentView = viewBudgetForm
		a.budgetForm.SetBudget(msg.Budget)
		return a, a.budgetForm.Init()
		
	case views.BackToDashboardMsg:
		a.currentView = viewDashboard
		return a, a.dashboard.Init()
	}

	switch a.currentView {
	case viewDashboard:
		a.dashboard, cmd = a.dashboard.Update(msg)
	case viewTransactions:
		a.transactionList, cmd = a.transactionList.Update(msg)
	case viewTransactionForm:
		a.transactionForm, cmd = a.transactionForm.Update(msg)
	case viewBudgets:
		a.budgetList, cmd = a.budgetList.Update(msg)
	case viewBudgetForm:
		a.budgetForm, cmd = a.budgetForm.Update(msg)
	case viewReports:
		a.reports, cmd = a.reports.Update(msg)
	case viewCategories:
		var model tea.Model
		model, cmd = a.categoryList.Update(msg)
		a.categoryList = model.(*views.CategoryListModel)
	case viewRecurring:
		var model tea.Model
		model, cmd = a.recurringList.Update(msg)
		a.recurringList = model.(*views.RecurringListModel)
		// Handle navigation back to dashboard on ESC/Q
		if msg, ok := msg.(tea.KeyMsg); ok && (msg.String() == "esc" || msg.String() == "q") {
			a.currentView = viewDashboard
			return a, a.dashboard.Init()
		}
	case viewRecurringForm:
		if a.recurringForm != nil {
			var model tea.Model
			model, cmd = a.recurringForm.Update(msg)
			a.recurringForm = model.(*views.RecurringFormModel)
			
			if a.recurringForm.IsCompleted() || a.recurringForm.IsCancelled() {
				a.currentView = viewRecurring
				return a, a.recurringList.Init()
			}
		}
	case viewCurrencySettings:
		a.currencySettings, cmd = a.currencySettings.Update(msg)
	}

	cmds = append(cmds, cmd)
	return a, tea.Batch(cmds...)
}

func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	var content string
	
	switch a.currentView {
	case viewDashboard:
		content = a.dashboard.View()
	case viewTransactions:
		content = a.transactionList.View()
	case viewTransactionForm:
		content = a.transactionForm.View()
	case viewBudgets:
		content = a.budgetList.View()
	case viewBudgetForm:
		content = a.budgetForm.View()
	case viewReports:
		content = a.reports.View()
	case viewCategories:
		content = a.categoryList.View()
	case viewRecurring:
		content = a.recurringList.View()
	case viewRecurringForm:
		if a.recurringForm != nil {
			content = a.recurringForm.View()
		}
	case viewCurrencySettings:
		content = a.currencySettings.View()
	}

	if a.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true).
			Padding(1)
		content += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", a.err))
	}

	return content
}

func (a *App) updateViewSizes() {
	if a.dashboard != nil {
		a.dashboard.SetSize(a.width, a.height)
	}
	if a.transactionList != nil {
		a.transactionList.SetSize(a.width, a.height)
	}
	if a.transactionForm != nil {
		a.transactionForm.SetSize(a.width, a.height)
	}
	if a.budgetList != nil {
		a.budgetList.SetSize(a.width, a.height)
	}
	if a.budgetForm != nil {
		a.budgetForm.SetSize(a.width, a.height)
	}
	if a.reports != nil {
		a.reports.SetSize(a.width, a.height)
	}
	if a.categoryList != nil {
		model, _ := a.categoryList.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
		a.categoryList = model.(*views.CategoryListModel)
	}
	if a.recurringList != nil {
		model, _ := a.recurringList.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
		a.recurringList = model.(*views.RecurringListModel)
	}
	if a.currencySettings != nil {
		a.currencySettings, _ = a.currencySettings.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
	}
}