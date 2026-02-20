package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"cloud.google.com/go/auth/credentials"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Spreadsheet interface {
	GetCurrentMonthWorksheet(ctx context.Context) (string, error)
	AddExpense(ctx context.Context, worksheet string, expense *Expense) error
	GetMonthlyTotal(ctx context.Context, worksheet string) (float64, error)
}

var _ Spreadsheet = (*SheetsService)(nil)

type SheetsService struct {
	service       *sheets.Service
	spreadsheetID string
	logger        *slog.Logger
}

func NewSheetsService(ctx context.Context, credentialsJSON, spreadsheetID string, logger *slog.Logger) (*SheetsService, error) {
	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: []string{
			"https://www.googleapis.com/auth/spreadsheets",
			"https://www.googleapis.com/auth/drive",
		},
		CredentialsJSON: []byte(credentialsJSON),
	})
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	service, err := sheets.NewService(ctx, option.WithAuthCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("create sheets service: %w", err)
	}

	return &SheetsService{
		service:       service,
		spreadsheetID: spreadsheetID,
		logger:        logger,
	}, nil
}

// GetCurrentMonthWorksheet returns the title of the newest worksheet
func (s *SheetsService) GetCurrentMonthWorksheet(ctx context.Context) (string, error) {
	spreadsheet, err := s.service.Spreadsheets.Get(s.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("get spreadsheet: %w", err)
	}

	if len(spreadsheet.Sheets) == 0 {
		return "", fmt.Errorf("no worksheets found")
	}

	return spreadsheet.Sheets[0].Properties.Title, nil
}

// AddExpense adds an expense to the given worksheet
func (s *SheetsService) AddExpense(ctx context.Context, worksheet string, expense *Expense) error {
	nextRow, err := s.findNextEmptyRow(ctx, worksheet)
	if err != nil {
		return fmt.Errorf("find next empty row: %w", err)
	}

	// Update cells: description in column A, amount in column B
	valueRange := &sheets.ValueRange{
		Values: [][]any{
			{expense.Desc, expense.Amount},
		},
	}

	rangeStr := fmt.Sprintf("%s!A%d:B%d", worksheet, nextRow, nextRow)
	_, err = s.service.Spreadsheets.Values.Update(s.spreadsheetID, rangeStr, valueRange).
		ValueInputOption("RAW").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("update cells: %w", err)
	}

	return nil
}

// GetMonthlyTotal calculates the total expenses for the current month
func (s *SheetsService) GetMonthlyTotal(ctx context.Context, worksheet string) (float64, error) {
	// Get columns for fundamentals and fun expenses
	colRanges := []string{
		fmt.Sprintf("%s!A:A", worksheet), // Fundamentals descriptions
		fmt.Sprintf("%s!B:B", worksheet), // Fundamentals amounts
		fmt.Sprintf("%s!C:C", worksheet), // Fun descriptions
		fmt.Sprintf("%s!D:D", worksheet), // Fun amounts
	}

	resp, err := s.service.Spreadsheets.Values.BatchGet(s.spreadsheetID).
		Ranges(colRanges...).
		Context(ctx).
		Do()
	if err != nil {
		return 0, fmt.Errorf("batch get values: %w", err)
	}

	if len(resp.ValueRanges) < 4 {
		return 0, fmt.Errorf("expected 4 value ranges, got %d", len(resp.ValueRanges))
	}

	return calculateMonthlyTotal(
		resp.ValueRanges[0].Values,
		resp.ValueRanges[1].Values,
		resp.ValueRanges[2].Values,
		resp.ValueRanges[3].Values,
	), nil
}

// Converts a Sheets API column into a []string
func flattenColumn(col [][]any) []string {
	result := make([]string, len(col))
	for i, row := range col {
		if len(row) > 0 {
			result[i] = strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		}
	}
	return result
}

func findExpenseStartRow(colValues []string) (int, bool) {
	for i, value := range colValues {
		if strings.Contains(value, "Total Net income") {
			// Expenses start after the header row following income
			return i + 2, true // Skip the header row
		}
	}
	return 0, false
}

func (s *SheetsService) findNextEmptyRow(ctx context.Context, worksheet string) (int, error) {
	rangeStr := fmt.Sprintf("%s!A:A", worksheet)
	resp, err := s.service.Spreadsheets.Values.Get(s.spreadsheetID, rangeStr).Context(ctx).Do()
	if err != nil {
		return 0, fmt.Errorf("get column values: %w", err)
	}

	colValues := flattenColumn(resp.Values)

	startRow, ok := findExpenseStartRow(colValues)
	if !ok {
		return 0, fmt.Errorf("could not find expense start row")
	}

	return nextEmptyRow(colValues, startRow), nil
}

// Finds the first empty row at or after startRow
// Returns a 1-indexed row number for the Sheets API
func nextEmptyRow(colValues []string, startRow int) int {
	for i := startRow; i < len(colValues)+1; i++ {
		if i > len(colValues) || colValues[i-1] == "" {
			return i + 1
		}
	}
	// If all rows are filled, append to the end
	return len(colValues) + 1
}

func calculateMonthlyTotal(fundamentalsDescRaw, fundamentalsAmountsRaw, funDescRaw, funAmountsRaw [][]any) float64 {
	fundamentalsDesc := flattenColumn(fundamentalsDescRaw)
	fundamentalsAmounts := flattenColumn(fundamentalsAmountsRaw)
	funDesc := flattenColumn(funDescRaw)
	funAmounts := flattenColumn(funAmountsRaw)

	startRow, ok := findExpenseStartRow(fundamentalsDesc)
	if !ok {
		return 0
	}

	fundamentalsTotal := sumColumnAmounts(fundamentalsAmounts, fundamentalsDesc, startRow)
	funTotal := sumColumnAmounts(funAmounts, funDesc, startRow)

	return fundamentalsTotal + funTotal
}

func sumColumnAmounts(amounts, descriptions []string, startRow int) float64 {
	total := 0.0

	for i := startRow - 1; i < len(amounts); i++ {
		if i < len(descriptions) && descriptions[i] != "" && amounts[i] != "" {
			amountStr := strings.ReplaceAll(amounts[i], ",", ".")
			if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
				total += amount
			}
		}
	}

	return total
}
