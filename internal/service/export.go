package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"budget-tracker/internal/models"
)

type ExportService struct {
	txService *TransactionService
}

func NewExportService(txService *TransactionService) *ExportService {
	return &ExportService{
		txService: txService,
	}
}

func (s *ExportService) ExportTransactionsCSV(writer io.Writer, filter *models.TransactionFilter) error {
	transactions, err := s.txService.GetByFilter(filter)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"Date",
		"Type",
		"Category",
		"Description",
		"Amount",
		"Currency",
		"Amount (USD)",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write transactions
	for _, tx := range transactions {
		record := []string{
			tx.Date.Format("2006-01-02"),
			string(tx.Type),
			tx.Category.Name,
			tx.Description,
			fmt.Sprintf("%.2f", tx.Amount),
			tx.Currency,
			fmt.Sprintf("%.2f", tx.AmountUSD),
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

func (s *ExportService) ExportMonthlyReportCSV(writer io.Writer, year int, month time.Month) error {
	summary, err := s.txService.GetMonthSummary(year, month)
	if err != nil {
		return fmt.Errorf("failed to get month summary: %w", err)
	}

	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	
	categoryTotals, err := s.txService.GetCategorySummary(start, end)
	if err != nil {
		return fmt.Errorf("failed to get category summary: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write report header
	if err := csvWriter.Write([]string{fmt.Sprintf("Monthly Report - %s %d", month.String(), year)}); err != nil {
		return err
	}
	if err := csvWriter.Write([]string{""}); err != nil {
		return err
	}

	// Write summary
	if err := csvWriter.Write([]string{"Summary"}); err != nil {
		return err
	}
	if err := csvWriter.Write([]string{"Total Income", fmt.Sprintf("%.2f", summary.TotalIncome)}); err != nil {
		return err
	}
	if err := csvWriter.Write([]string{"Total Expenses", fmt.Sprintf("%.2f", summary.TotalExpenses)}); err != nil {
		return err
	}
	if err := csvWriter.Write([]string{"Balance", fmt.Sprintf("%.2f", summary.Balance)}); err != nil {
		return err
	}
	if err := csvWriter.Write([]string{""}); err != nil {
		return err
	}

	// Write category breakdown
	if err := csvWriter.Write([]string{"Category Breakdown"}); err != nil {
		return err
	}
	if err := csvWriter.Write([]string{"Category", "Type", "Total", "Count", "Percentage"}); err != nil {
		return err
	}

	for _, cat := range categoryTotals {
		record := []string{
			cat.Name,
			string(cat.Type),
			fmt.Sprintf("%.2f", cat.Total),
			fmt.Sprintf("%d", cat.Count),
			fmt.Sprintf("%.1f%%", cat.Percentage),
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func (s *ExportService) ExportBudgetStatusCSV(writer io.Writer, budgetService *BudgetService) error {
	statuses, err := budgetService.GetAllStatuses()
	if err != nil {
		return fmt.Errorf("failed to get budget statuses: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"Budget Name",
		"Category",
		"Period",
		"Budget Amount",
		"Spent",
		"Remaining",
		"Percent Used",
		"Status",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write budget statuses
	for _, status := range statuses {
		statusText := "OK"
		if status.IsOverBudget {
			statusText = "OVER BUDGET"
		}

		record := []string{
			status.Budget.Name,
			status.Budget.Category.Name,
			string(status.Budget.Period),
			fmt.Sprintf("%.2f", status.Budget.Amount),
			fmt.Sprintf("%.2f", status.Spent),
			fmt.Sprintf("%.2f", status.Remaining),
			fmt.Sprintf("%.1f%%", status.PercentUsed),
			statusText,
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}