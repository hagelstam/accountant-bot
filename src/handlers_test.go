package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type mockSender struct {
	sendFunc func(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
	calls    []*bot.SendMessageParams
}

func (m *mockSender) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	m.calls = append(m.calls, params)
	if m.sendFunc != nil {
		return m.sendFunc(ctx, params)
	}
	return &models.Message{}, nil
}

var _ Spreadsheet = (*mockSheet)(nil)

type mockSheet struct {
	getWorksheetFunc func(ctx context.Context) (string, error)
	addExpenseFunc   func(ctx context.Context, worksheet string, expense *Expense) error
	getMonthlyFunc   func(ctx context.Context, worksheet string) (float64, error)
}

func (m *mockSheet) GetCurrentMonthWorksheet(ctx context.Context) (string, error) {
	if m.getWorksheetFunc != nil {
		return m.getWorksheetFunc(ctx)
	}
	return "February 2026", nil
}

func (m *mockSheet) AddExpense(ctx context.Context, worksheet string, expense *Expense) error {
	if m.addExpenseFunc != nil {
		return m.addExpenseFunc(ctx, worksheet, expense)
	}
	return nil
}

func (m *mockSheet) GetMonthlyTotal(ctx context.Context, worksheet string) (float64, error) {
	if m.getMonthlyFunc != nil {
		return m.getMonthlyFunc(ctx, worksheet)
	}
	return 100.0, nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestFormatAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount float64
		want   string
	}{
		{10.0, "10,00"},
		{12.95, "12,95"},
		{12.956, "12,96"},
		{0.0, "0,00"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := formatAmount(tt.amount); got != tt.want {
				t.Errorf("formatAmount(%v) = %q, want %q", tt.amount, got, tt.want)
			}
		})
	}
}

func TestHandleStart(t *testing.T) {
	t.Parallel()

	t.Run("sends welcome message with user name", func(t *testing.T) {
		t.Parallel()

		sender := &mockSender{}
		h := NewBotHandlers(&mockSheet{}, discardLogger())

		update := &models.Update{
			Message: &models.Message{
				Chat: models.Chat{ID: 123},
				From: &models.User{FirstName: "Alice"},
			},
		}

		h.HandleStart(context.Background(), sender, update)

		if len(sender.calls) != 1 {
			t.Fatalf("expected 1 SendMessage call, got %d", len(sender.calls))
		}
		if !strings.Contains(sender.calls[0].Text, "Hi Alice!") {
			t.Errorf("welcome message should contain user's first name, got %q", sender.calls[0].Text)
		}
	})

	t.Run("nil message is no-op", func(t *testing.T) {
		t.Parallel()

		sender := &mockSender{}
		h := NewBotHandlers(&mockSheet{}, discardLogger())

		h.HandleStart(context.Background(), sender, &models.Update{Message: nil})

		if len(sender.calls) != 0 {
			t.Errorf("expected 0 SendMessage calls, got %d", len(sender.calls))
		}
	})
}

func TestHandleExpense(t *testing.T) {
	t.Parallel()

	errFunc := func(ctx context.Context) (string, error) {
		return "", fmt.Errorf("fail")
	}

	tests := []struct {
		name         string
		update       *models.Update
		sheet        *mockSheet
		wantErr      bool
		wantCalls    int
		wantContains []string
	}{
		{
			name:   "nil message",
			update: &models.Update{Message: nil},
			sheet:  &mockSheet{},
		},
		{
			name: "invalid format sends parse error",
			update: &models.Update{
				Message: &models.Message{Chat: models.Chat{ID: 1}, Text: "invalid"},
			},
			sheet:        &mockSheet{},
			wantCalls:    1,
			wantContains: []string{"Could not parse expense"},
		},
		{
			name: "valid expense",
			update: &models.Update{
				Message: &models.Message{Chat: models.Chat{ID: 1}, Text: "Lunch 12,50"},
			},
			sheet: &mockSheet{
				getMonthlyFunc: func(ctx context.Context, worksheet string) (float64, error) {
					return 150.50, nil
				},
			},
			wantCalls:    1,
			wantContains: []string{"12,50", "Lunch", "150,50"},
		},
		{
			name: "worksheet error returns error",
			update: &models.Update{
				Message: &models.Message{Chat: models.Chat{ID: 1}, Text: "Lunch 12.50"},
			},
			sheet:   &mockSheet{getWorksheetFunc: errFunc},
			wantErr: true,
		},
		{
			name: "add expense error returns error",
			update: &models.Update{
				Message: &models.Message{Chat: models.Chat{ID: 1}, Text: "Lunch 12.50"},
			},
			sheet: &mockSheet{addExpenseFunc: func(ctx context.Context, ws string, e *Expense) error {
				return fmt.Errorf("fail")
			}},
			wantErr: true,
		},
		{
			name: "monthly total error still sends response",
			update: &models.Update{
				Message: &models.Message{Chat: models.Chat{ID: 1}, Text: "Lunch 12.50"},
			},
			sheet: &mockSheet{getMonthlyFunc: func(ctx context.Context, ws string) (float64, error) {
				return 0, fmt.Errorf("fail")
			}},
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sender := &mockSender{}
			h := NewBotHandlers(tt.sheet, discardLogger())

			err := h.HandleExpense(context.Background(), sender, tt.update)

			if (err != nil) != tt.wantErr {
				t.Fatalf("HandleExpense() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(sender.calls) != tt.wantCalls {
				t.Fatalf("expected %d SendMessage calls, got %d", tt.wantCalls, len(sender.calls))
			}
			for _, s := range tt.wantContains {
				if !strings.Contains(sender.calls[0].Text, s) {
					t.Errorf("response should contain %q, got %q", s, sender.calls[0].Text)
				}
			}
		})
	}
}
