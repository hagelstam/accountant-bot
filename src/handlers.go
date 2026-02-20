package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Sender interface {
	SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
}

type BotHandlers struct {
	sheets Spreadsheet
	logger *slog.Logger
}

func NewBotHandlers(sheets Spreadsheet, logger *slog.Logger) *BotHandlers {
	return &BotHandlers{
		sheets: sheets,
		logger: logger,
	}
}

// HandleStart handles the /start command
func (h *BotHandlers) HandleStart(ctx context.Context, sender Sender, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		return
	}

	user := update.Message.From
	h.logger.Info("user started bot",
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Username))

	welcomeMessage := fmt.Sprintf(
		"Hi %s! 👋\n\n"+
			"I'm your personal accountant bot. Send me expenses in this format:\n\n"+
			"Example: `Lunch 2.95`\n\n",
		user.FirstName,
	)

	_, err := sender.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   welcomeMessage,
	})
	if err != nil {
		h.logger.Error("failed to send welcome message", slog.String("error", err.Error()))
	}
}

// HandleExpense handles expense messages
// Returns an error only for sheet update failures that should trigger an SQS retry
func (h *BotHandlers) HandleExpense(ctx context.Context, sender Sender, update *models.Update) error {
	if update.Message == nil || update.Message.Text == "" {
		return nil
	}

	messageText := update.Message.Text
	expense, err := ParseExpense(messageText)
	if err != nil {
		_, sendErr := sender.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Could not parse expense. Please use format:\n\nExample: `Lunch 2.95`",
		})
		if sendErr != nil {
			h.logger.Error("failed to send error message", slog.String("error", sendErr.Error()))
		}
		return nil
	}

	worksheet, err := h.sheets.GetCurrentMonthWorksheet(ctx)
	if err != nil {
		return fmt.Errorf("get worksheet: %w", err)
	}

	if err := h.sheets.AddExpense(ctx, worksheet, expense); err != nil {
		return fmt.Errorf("add expense: %w", err)
	}

	monthlyTotal, err := h.sheets.GetMonthlyTotal(ctx, worksheet)
	if err != nil {
		h.logger.Error("failed to get monthly total", slog.String("error", err.Error()))
	}

	response := fmt.Sprintf(
		"💸 Spent %s€ on %s. New monthly total is %s€",
		formatAmount(expense.Amount),
		expense.Desc,
		formatAmount(monthlyTotal),
	)

	_, err = sender.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   response,
	})
	if err != nil {
		h.logger.Error("failed to send response message", slog.String("error", err.Error()))
	}

	return nil
}

func formatAmount(amount float64) string {
	formatted := fmt.Sprintf("%.2f", amount)
	return strings.ReplaceAll(formatted, ".", ",")
}
