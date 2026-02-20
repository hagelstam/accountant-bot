package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var expensePattern = regexp.MustCompile(`^(.+?)\s+([\d,.]+)$`)

type Expense struct {
	Desc   string
	Amount float64
}

// ParseExpense parses an expense from a message in the format "<Desc> <Amount>"
// Example message: "Lunch 2.95"
func ParseExpense(message string) (*Expense, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("empty message")
	}

	matches := expensePattern.FindStringSubmatch(message)
	if matches == nil {
		return nil, fmt.Errorf("invalid expense format")
	}

	amountStr := strings.ReplaceAll(matches[2], ",", ".")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("parse amount: %w", err)
	}

	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	return &Expense{
		Desc:   strings.TrimSpace(matches[1]),
		Amount: amount,
	}, nil
}
