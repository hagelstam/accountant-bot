package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-telegram/bot/models"
)

func newTestApp(sender *mockSender, sheet *mockSheet) *app {
	logger := discardLogger()
	return &app{
		sender:   sender,
		handlers: NewBotHandlers(sheet, logger),
		logger:   logger,
	}
}

func TestProcessUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		update    *models.Update
		wantCalls int
	}{
		{
			name:   "nil message",
			update: &models.Update{Message: nil},
		},
		{
			name: "start command",
			update: &models.Update{
				Message: &models.Message{
					Chat: models.Chat{ID: 1},
					From: &models.User{FirstName: "Test"},
					Text: "/start",
				},
			},
			wantCalls: 1,
		},
		{
			name: "expense message",
			update: &models.Update{
				Message: &models.Message{Chat: models.Chat{ID: 1}, Text: "Lunch 12.50"},
			},
			wantCalls: 1,
		},
		{
			name: "empty text",
			update: &models.Update{
				Message: &models.Message{Chat: models.Chat{ID: 1}, Text: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &mockSender{}
			a := newTestApp(s, &mockSheet{})

			err := a.processUpdate(context.Background(), tt.update)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(s.calls) != tt.wantCalls {
				t.Errorf("expected %d SendMessage calls, got %d", tt.wantCalls, len(s.calls))
			}
		})
	}
}

func TestHandleRequest(t *testing.T) {
	t.Parallel()

	mustMarshal := func(v any) string {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		return string(b)
	}

	t.Run("valid record", func(t *testing.T) {
		t.Parallel()

		a := newTestApp(&mockSender{}, &mockSheet{})

		body := mustMarshal(models.Update{
			ID:      1,
			Message: &models.Message{Chat: models.Chat{ID: 1}, Text: "Lunch 12.50"},
		})
		event := events.SQSEvent{
			Records: []events.SQSMessage{{MessageId: "msg-1", Body: body}},
		}

		resp, err := a.handleRequest(context.Background(), event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.BatchItemFailures) != 0 {
			t.Errorf("expected 0 failures, got %d", len(resp.BatchItemFailures))
		}
	})

	t.Run("invalid json skipped", func(t *testing.T) {
		t.Parallel()

		a := newTestApp(&mockSender{}, &mockSheet{})

		event := events.SQSEvent{
			Records: []events.SQSMessage{{MessageId: "msg-1", Body: "not json"}},
		}

		resp, err := a.handleRequest(context.Background(), event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.BatchItemFailures) != 0 {
			t.Errorf("expected 0 failures, got %d", len(resp.BatchItemFailures))
		}
	})

	t.Run("process error adds to failures", func(t *testing.T) {
		t.Parallel()

		sheet := &mockSheet{
			getWorksheetFunc: func(ctx context.Context) (string, error) {
				return "", fmt.Errorf("fail")
			},
		}
		a := newTestApp(&mockSender{}, sheet)

		body := mustMarshal(models.Update{
			ID:      1,
			Message: &models.Message{Chat: models.Chat{ID: 1}, Text: "Lunch 12.50"},
		})
		event := events.SQSEvent{
			Records: []events.SQSMessage{{MessageId: "msg-1", Body: body}},
		}

		resp, err := a.handleRequest(context.Background(), event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.BatchItemFailures) != 1 {
			t.Fatalf("expected 1 failure, got %d", len(resp.BatchItemFailures))
		}
		if resp.BatchItemFailures[0].ItemIdentifier != "msg-1" {
			t.Errorf("failure identifier = %q, want %q", resp.BatchItemFailures[0].ItemIdentifier, "msg-1")
		}
	})
}
