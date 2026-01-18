package main

import (
	"testing"
)

func TestFormatAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		amount float64
		want   string
	}{
		{
			name:   "whole number",
			amount: 10.0,
			want:   "10,00",
		},
		{
			name:   "two decimals",
			amount: 12.95,
			want:   "12,95",
		},
		{
			name:   "rounding up",
			amount: 12.956,
			want:   "12,96",
		},
		{
			name:   "rounding down",
			amount: 12.954,
			want:   "12,95",
		},
		{
			name:   "zero",
			amount: 0.0,
			want:   "0,00",
		},
		{
			name:   "large amount",
			amount: 1234.56,
			want:   "1234,56",
		},
		{
			name:   "one decimal place",
			amount: 5.5,
			want:   "5,50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatAmount(tt.amount)
			if got != tt.want {
				t.Errorf("formatAmount(%v) = %v, want %v", tt.amount, got, tt.want)
			}
		})
	}
}
