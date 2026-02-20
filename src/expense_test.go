package main

import (
	"testing"
)

func TestParseExpense(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantDesc   string
		wantAmount float64
		wantErr    bool
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
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "missing amount",
			input:   "Lunch",
			wantErr: true,
		},
		{
			name:    "invalid amount",
			input:   "Lunch abc",
			wantErr: true,
		},
		{
			name:    "negative amount",
			input:   "Lunch -2.95",
			wantErr: true,
		},
		{
			name:    "zero amount",
			input:   "Lunch 0",
			wantErr: true,
		},
		{
			name:    "zero amount with decimals",
			input:   "Lunch 0.00",
			wantErr: true,
		},
		{
			name:    "unparseable number matching regex",
			input:   "Lunch 1.2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseExpense(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseExpense() expected error, got nil")
				}
				if result != nil {
					t.Errorf("ParseExpense() = %v, want nil when error", result)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseExpense() unexpected error = %v", err)
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
