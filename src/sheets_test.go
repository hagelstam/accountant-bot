package main

import (
	"testing"
)

func TestFlattenColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		col  [][]any
		want []string
	}{
		{
			name: "normal values",
			col:  [][]any{{"hello"}, {"world"}},
			want: []string{"hello", "world"},
		},
		{
			name: "empty rows",
			col:  [][]any{{}, {"value"}, {}},
			want: []string{"", "value", ""},
		},
		{
			name: "nil input",
			col:  nil,
			want: []string{},
		},
		{
			name: "numeric values",
			col:  [][]any{{42}, {3.14}},
			want: []string{"42", "3.14"},
		},
		{
			name: "whitespace trimming",
			col:  [][]any{{"  hello  "}, {" world "}},
			want: []string{"hello", "world"},
		},
		{
			name: "multiple values per row uses first",
			col:  [][]any{{"first", "second"}, {"a", "b"}},
			want: []string{"first", "a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := flattenColumn(tt.col)

			if len(got) != len(tt.want) {
				t.Fatalf("flattenColumn() length = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("flattenColumn()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFindExpenseStartRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		colValues []string
		wantRow   int
		wantFound bool
	}{
		{
			name:      "found in middle",
			colValues: []string{"Income", "Salary", "Total Net income", "Expenses", "Rent"},
			wantRow:   4, // index 2 + 2
			wantFound: true,
		},
		{
			name:      "found at start",
			colValues: []string{"Total Net income", "Header", "First expense"},
			wantRow:   2, // index 0 + 2
			wantFound: true,
		},
		{
			name:      "not found",
			colValues: []string{"Income", "Salary", "Other"},
			wantRow:   0,
			wantFound: false,
		},
		{
			name:      "empty input",
			colValues: []string{},
			wantRow:   0,
			wantFound: false,
		},
		{
			name:      "partial match",
			colValues: []string{"Some Total Net income row", "Header", "Expense"},
			wantRow:   2, // index 0 + 2 (Contains match)
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			row, found := findExpenseStartRow(tt.colValues)

			if found != tt.wantFound {
				t.Errorf("findExpenseStartRow() found = %v, want %v", found, tt.wantFound)
			}
			if row != tt.wantRow {
				t.Errorf("findExpenseStartRow() row = %d, want %d", row, tt.wantRow)
			}
		})
	}
}

func TestNextEmptyRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		colValues []string
		startRow  int
		want      int
	}{
		{
			name:      "first row empty",
			colValues: []string{"Total Net income", "Header", ""},
			startRow:  3,
			want:      4, // index 2 is empty, so row 3+1=4 in sheets API
		},
		{
			name:      "gap in middle",
			colValues: []string{"Total Net income", "Header", "Rent", "", "Coffee"},
			startRow:  3,
			want:      5, // index 3 is empty, so row 4+1=5
		},
		{
			name:      "all rows filled",
			colValues: []string{"Total Net income", "Header", "Rent", "Food"},
			startRow:  3,
			want:      5, // appends after last row
		},
		{
			name:      "startRow beyond data",
			colValues: []string{"Total Net income", "Header"},
			startRow:  3,
			want:      3, // loop doesn't execute, returns len+1=3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := nextEmptyRow(tt.colValues, tt.startRow)
			if got != tt.want {
				t.Errorf("nextEmptyRow() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateMonthlyTotal(t *testing.T) {
	t.Parallel()

	t.Run("sums fundamentals and fun", func(t *testing.T) {
		t.Parallel()

		fundDesc := [][]any{{"Income"}, {"Total Net income"}, {"Expenses"}, {"Rent"}, {"Food"}}
		fundAmounts := [][]any{{""}, {""}, {""}, {"500.00"}, {"100.00"}}
		funDesc := [][]any{{""}, {""}, {""}, {"Movies"}, {"Games"}}
		funAmounts := [][]any{{""}, {""}, {""}, {"15.00"}, {"30.00"}}

		got := calculateMonthlyTotal(fundDesc, fundAmounts, funDesc, funAmounts)
		want := 645.0
		if got != want {
			t.Errorf("calculateMonthlyTotal() = %v, want %v", got, want)
		}
	})

	t.Run("no expense start row returns zero", func(t *testing.T) {
		t.Parallel()

		fundDesc := [][]any{{"Income"}, {"Salary"}}
		fundAmounts := [][]any{{"1000"}, {"2000"}}
		funDesc := [][]any{{}, {}}
		funAmounts := [][]any{{}, {}}

		got := calculateMonthlyTotal(fundDesc, fundAmounts, funDesc, funAmounts)
		if got != 0 {
			t.Errorf("calculateMonthlyTotal() = %v, want 0", got)
		}
	})

	t.Run("empty columns", func(t *testing.T) {
		t.Parallel()

		got := calculateMonthlyTotal(nil, nil, nil, nil)
		if got != 0 {
			t.Errorf("calculateMonthlyTotal() = %v, want 0", got)
		}
	})
}

func TestSumColumnAmounts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		amounts      []string
		descriptions []string
		startRow     int
		want         float64
	}{
		{
			name:         "basic sum",
			amounts:      []string{"", "", "10.00", "20.50", "5.00"},
			descriptions: []string{"", "", "Rent", "Food", "Coffee"},
			startRow:     3,
			want:         35.50,
		},
		{
			name:         "comma decimal separator",
			amounts:      []string{"10,50", "20,00"},
			descriptions: []string{"Rent", "Food"},
			startRow:     1,
			want:         30.50,
		},
		{
			name:         "skips empty descriptions",
			amounts:      []string{"10.00", "20.00", "30.00"},
			descriptions: []string{"Rent", "", "Food"},
			startRow:     1,
			want:         40.00,
		},
		{
			name:         "skips empty amounts",
			amounts:      []string{"10.00", ""},
			descriptions: []string{"Rent", "Food"},
			startRow:     1,
			want:         10.00,
		},
		{
			name:         "empty input",
			amounts:      []string{},
			descriptions: []string{},
			startRow:     1,
			want:         0,
		},
		{
			name:         "skips unparseable amounts",
			amounts:      []string{"10.00", "abc", "20.00"},
			descriptions: []string{"Rent", "Bad", "Food"},
			startRow:     1,
			want:         30.00,
		},
		{
			name:         "amounts shorter than descriptions",
			amounts:      []string{"10.00"},
			descriptions: []string{"Rent", "Food", "Coffee"},
			startRow:     1,
			want:         10.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := sumColumnAmounts(tt.amounts, tt.descriptions, tt.startRow)
			if got != tt.want {
				t.Errorf("sumColumnAmounts() = %v, want %v", got, tt.want)
			}
		})
	}
}
