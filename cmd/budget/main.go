package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"budget-tracker/internal/db"
	"budget-tracker/internal/models"
	"budget-tracker/internal/repository"
	"budget-tracker/internal/service"
	"budget-tracker/internal/ui"
)

func main() {
	// Parse command-line flags
	exportCmd := flag.String("export", "", "Export data to CSV (transactions, report, budgets)")
	outputFile := flag.String("output", "", "Output file for export")
	monthFlag := flag.Int("month", 0, "Month for report export (1-12)")
	yearFlag := flag.Int("year", time.Now().Year(), "Year for report export")
	flag.Parse()

	// Handle export command
	if *exportCmd != "" {
		handleExport(*exportCmd, *outputFile, *monthFlag, *yearFlag)
		return
	}
	database, err := db.InitDB(db.GetDefaultDBPath())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	defer sqlDB.Close()

	// Initialize settings service
	settingsService, err := service.NewSettingsService("data")
	if err != nil {
		log.Fatalf("Failed to initialize settings: %v", err)
	}

	txRepo := repository.NewTransactionRepository(database)
	categoryRepo := repository.NewCategoryRepository(database)
	budgetRepo := repository.NewBudgetRepository(database)
	recurringRepo := repository.NewRecurringTransactionRepository(database)

	currencyService := service.NewCurrencyService(settingsService)
	txService := service.NewTransactionService(txRepo, currencyService)
	categoryService := service.NewCategoryService(categoryRepo)
	budgetService := service.NewBudgetService(budgetRepo, txRepo)
	recurringService := service.NewRecurringTransactionService(recurringRepo, txRepo, currencyService)

	// Process any due recurring transactions on startup
	if _, err := recurringService.ProcessDueTransactions(time.Now()); err != nil {
		log.Printf("Warning: Failed to process recurring transactions: %v", err)
	}

	app := ui.NewApp(txService, categoryService, budgetService, currencyService, settingsService, recurringService)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func handleExport(exportType, outputFile string, month, year int) {
	database, err := db.InitDB(db.GetDefaultDBPath())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	defer sqlDB.Close()

	// Initialize settings service
	settingsService, err := service.NewSettingsService("data")
	if err != nil {
		log.Fatalf("Failed to initialize settings: %v", err)
	}

	// Initialize services
	txRepo := repository.NewTransactionRepository(database)
	budgetRepo := repository.NewBudgetRepository(database)
	
	currencyService := service.NewCurrencyService(settingsService)
	txService := service.NewTransactionService(txRepo, currencyService)
	budgetService := service.NewBudgetService(budgetRepo, txRepo)
	exportService := service.NewExportService(txService)

	// Determine output
	var output *os.File
	if outputFile == "" {
		output = os.Stdout
	} else {
		output, err = os.Create(outputFile)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer output.Close()
	}

	switch exportType {
	case "transactions":
		filter := &models.TransactionFilter{}
		if err := exportService.ExportTransactionsCSV(output, filter); err != nil {
			log.Fatalf("Failed to export transactions: %v", err)
		}
		if outputFile != "" {
			fmt.Printf("Transactions exported to %s\n", outputFile)
		}

	case "report":
		if month == 0 {
			month = int(time.Now().Month())
		}
		if err := exportService.ExportMonthlyReportCSV(output, year, time.Month(month)); err != nil {
			log.Fatalf("Failed to export report: %v", err)
		}
		if outputFile != "" {
			fmt.Printf("Monthly report exported to %s\n", outputFile)
		}

	case "budgets":
		if err := exportService.ExportBudgetStatusCSV(output, budgetService); err != nil {
			log.Fatalf("Failed to export budgets: %v", err)
		}
		if outputFile != "" {
			fmt.Printf("Budget status exported to %s\n", outputFile)
		}

	default:
		fmt.Printf("Unknown export type: %s\n", exportType)
		fmt.Println("Available types: transactions, report, budgets")
		os.Exit(1)
	}
}