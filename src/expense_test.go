package main

import (
	"testing"
)

func TestParseExpense(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantDesc   string
		wantAmount float64
		wantNil    bool
	}{
		{
			name:       "simple expense",
			input:      "Lunch 2.95",
			wantDesc:   "Lunch",
			wantAmount: 2.95,
		},
		{
			name:       "expense with multiple words",
			input:      "Gym membership 31.99",
			wantDesc:   "Gym membership",
			wantAmount: 31.99,
		},
		{
			name:       "expense with comma decimal",
			input:      "Groceries 15,50",
			wantDesc:   "Groceries",
			wantAmount: 15.5,
		},
		{
			name:       "expense with whole number",
			input:      "Movie ticket 12",
			wantDesc:   "Movie ticket",
			wantAmount: 12.0,
		},
		{
			name:       "expense with extra whitespace",
			input:      "  Lunch   2.95  ",
			wantDesc:   "Lunch",
			wantAmount: 2.95,
		},
		{
			name:    "empty string",
			input:   "",
			wantNil: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantNil: true,
		},
		{
			name:    "missing amount",
			input:   "Lunch",
			wantNil: true,
		},
		{
			name:    "invalid amount",
			input:   "Lunch abc",
			wantNil: true,
		},
		{
			name:    "negative amount",
			input:   "Lunch -2.95",
			wantNil: true,
		},
		{
			name:    "zero amount",
			input:   "Lunch 0",
			wantNil: true,
		},
		{
			name:    "zero amount with decimals",
			input:   "Lunch 0.00",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseExpense(tt.input)

			if err != nil {
				t.Errorf("ParseExpense() unexpected error = %v", err)
				return
			}

			if tt.wantNil {
				if result != nil {
					t.Errorf("ParseExpense() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Errorf("ParseExpense() = nil, want non-nil")
				return
			}

			if result.Desc != tt.wantDesc {
				t.Errorf("ParseExpense().Desc = %v, want %v", result.Desc, tt.wantDesc)
			}

			if result.Amount != tt.wantAmount {
				t.Errorf("ParseExpense().Amount = %v, want %v", result.Amount, tt.wantAmount)
			}
		})
	}
}
