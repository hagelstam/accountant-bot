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
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	srv, err := sheets.NewService(ctx, option.WithAuthCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &SheetsService{
		service:       srv,
		spreadsheetID: spreadsheetID,
		logger:        logger,
	}, nil
}

// AddExpense adds an expense to the current month's worksheet
func (s *SheetsService) AddExpense(ctx context.Context, expense *Expense) error {
	worksheet, err := s.getCurrentMonthWorksheet(ctx)
	if err != nil {
		return fmt.Errorf("get current worksheet: %w", err)
	}

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
func (s *SheetsService) GetMonthlyTotal(ctx context.Context) (float64, error) {
	worksheet, err := s.getCurrentMonthWorksheet(ctx)
	if err != nil {
		return 0, fmt.Errorf("get current worksheet: %w", err)
	}

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
		return 0, nil
	}

	fundamentalsDesc := resp.ValueRanges[0].Values
	fundamentalsAmounts := resp.ValueRanges[1].Values
	funDesc := resp.ValueRanges[2].Values
	funAmounts := resp.ValueRanges[3].Values

	startRow := s.findExpenseStartRow(fundamentalsDesc)
	if startRow == -1 {
		return 0, nil
	}

	fundamentalsTotal := s.sumColumnAmounts(fundamentalsAmounts, fundamentalsDesc, startRow)
	funTotal := s.sumColumnAmounts(funAmounts, funDesc, startRow)

	return fundamentalsTotal + funTotal, nil
}

func (s *SheetsService) getCurrentMonthWorksheet(ctx context.Context) (string, error) {
	spreadsheet, err := s.service.Spreadsheets.Get(s.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("get spreadsheet: %w", err)
	}

	if len(spreadsheet.Sheets) == 0 {
		return "", fmt.Errorf("no worksheets found")
	}

	// The leftmost (newest) sheet is the first in the list
	return spreadsheet.Sheets[0].Properties.Title, nil
}

func (s *SheetsService) findExpenseStartRow(colValues [][]interface{}) int {
	for i, row := range colValues {
		if len(row) > 0 {
			value := fmt.Sprintf("%v", row[0])
			if strings.Contains(value, "Total Net income") {
				// Expenses start after the header row following income
				return i + 2 // Skip the header row (0-indexed, so +2 gives us the row after next)
			}
		}
	}
	return -1
}

func (s *SheetsService) findNextEmptyRow(ctx context.Context, worksheet string) (int, error) {
	rangeStr := fmt.Sprintf("%s!A:A", worksheet)
	resp, err := s.service.Spreadsheets.Values.Get(s.spreadsheetID, rangeStr).Context(ctx).Do()
	if err != nil {
		return 0, fmt.Errorf("get column values: %w", err)
	}

	colValues := resp.Values

	startRow := s.findExpenseStartRow(colValues)
	if startRow == -1 {
		return 0, fmt.Errorf("could not find expense start row")
	}

	// Find the next empty row after startRow (convert to 1-indexed)
	for i := startRow; i < len(colValues)+1; i++ {
		if i > len(colValues) || len(colValues[i-1]) == 0 || strings.TrimSpace(fmt.Sprintf("%v", colValues[i-1][0])) == "" {
			return i + 1, nil // +1 for 1-indexed sheets API
		}
	}

	// If all rows are filled, append to the end
	return len(colValues) + 1, nil
}

func (s *SheetsService) sumColumnAmounts(amounts, descriptions [][]interface{}, startRow int) float64 {
	total := 0.0

	for i := startRow - 1; i < len(amounts); i++ {
		// Check if description exists and is not empty
		if i < len(descriptions) && len(descriptions[i]) > 0 {
			desc := strings.TrimSpace(fmt.Sprintf("%v", descriptions[i][0]))
			if desc != "" && len(amounts[i]) > 0 {
				amountStr := fmt.Sprintf("%v", amounts[i][0])
				amountStr = strings.ReplaceAll(amountStr, ",", ".")
				if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
					total += amount
				}
			}
		}
	}

	return total
}
