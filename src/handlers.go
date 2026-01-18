package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type BotHandlers struct {
	sheets *SheetsService
	logger *slog.Logger
}

func NewBotHandlers(sheets *SheetsService, logger *slog.Logger) *BotHandlers {
	return &BotHandlers{
		sheets: sheets,
		logger: logger,
	}
}

// HandleStart handles the /start command
func (h *BotHandlers) HandleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
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

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   welcomeMessage,
	})
	if err != nil {
		h.logger.Error("failed to send welcome message", slog.String("error", err.Error()))
	}
}

// HandleExpense handles expense messages
func (h *BotHandlers) HandleExpense(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	messageText := update.Message.Text
	h.logger.Info("received message", slog.String("text", messageText))

	expense, err := ParseExpense(messageText)
	if err != nil || expense == nil {
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Could not parse expense. Please use format:\n\nExample: `Lunch 2.95`",
		})
		if sendErr != nil {
			h.logger.Error("failed to send error message", slog.String("error", sendErr.Error()))
		}
		return
	}

	if err := h.sheets.AddExpense(ctx, expense); err != nil {
		h.logger.Error("failed to add expense", slog.String("error", err.Error()))
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Failed to add expense: %v", err),
		})
		if sendErr != nil {
			h.logger.Error("failed to send error message", slog.String("error", sendErr.Error()))
		}
		return
	}

	monthlyTotal, err := h.sheets.GetMonthlyTotal(ctx)
	if err != nil {
		h.logger.Error("failed to get monthly total", slog.String("error", err.Error()))
	}

	formattedAmount := formatAmount(expense.Amount)
	formattedTotal := formatAmount(monthlyTotal)

	response := fmt.Sprintf(
		"💸 Spent %s€ on %s. New monthly total is %s€",
		formattedAmount,
		expense.Desc,
		formattedTotal,
	)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   response,
	})
	if err != nil {
		h.logger.Error("failed to send response message", slog.String("error", err.Error()))
	}
}

func formatAmount(amount float64) string {
	formatted := fmt.Sprintf("%.2f", amount)
	return strings.ReplaceAll(formatted, ".", ",")
}
