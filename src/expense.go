package main

import (
	"regexp"
	"strconv"
	"strings"
)

type Expense struct {
	Desc   string
	Amount float64
}

// ParseExpense parses an expense from a message in the format "Description Amount"
// Example message: "Lunch 2.95"
func ParseExpense(message string) (*Expense, error) {
	if strings.TrimSpace(message) == "" {
		return nil, nil
	}

	pattern := regexp.MustCompile(`^(.+?)\s+([\d,.]+)$`)
	matches := pattern.FindStringSubmatch(strings.TrimSpace(message))

	if matches == nil {
		return nil, nil
	}

	desc := strings.TrimSpace(matches[1])
	amountStr := strings.ReplaceAll(matches[2], ",", ".")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, nil
	}

	if amount <= 0 {
		return nil, nil
	}

	return &Expense{
		Desc:   desc,
		Amount: amount,
	}, nil
}
