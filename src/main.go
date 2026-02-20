package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type app struct {
	sender   Sender
	handlers *BotHandlers
	logger   *slog.Logger
}

func newApp() (*app, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.LogLevel,
	}))

	ctx := context.Background()
	sheetsService, err := NewSheetsService(ctx, config.GoogleCredentialsJSON, config.GoogleSpreadsheetID, logger)
	if err != nil {
		return nil, fmt.Errorf("create sheets service: %w", err)
	}

	handlers := NewBotHandlers(sheetsService, logger)

	telegramBot, err := bot.New(config.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	return &app{
		sender:   telegramBot,
		handlers: handlers,
		logger:   logger,
	}, nil
}

func (a *app) handleRequest(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	var failures []events.SQSBatchItemFailure

	for _, record := range event.Records {
		var update models.Update
		if err := json.Unmarshal([]byte(record.Body), &update); err != nil {
			a.logger.Error("failed to unmarshal update",
				slog.String("message_id", record.MessageId),
				slog.String("error", err.Error()))
			continue
		}

		if err := a.processUpdate(ctx, &update); err != nil {
			a.logger.Error("failed to process update",
				slog.String("message_id", record.MessageId),
				slog.Int64("update_id", update.ID),
				slog.String("error", err.Error()))
			failures = append(failures, events.SQSBatchItemFailure{
				ItemIdentifier: record.MessageId,
			})
			continue
		}

		a.logger.Info("successfully processed update",
			slog.String("message_id", record.MessageId),
			slog.Int64("update_id", update.ID))
	}

	return events.SQSEventResponse{BatchItemFailures: failures}, nil
}

func (a *app) processUpdate(ctx context.Context, update *models.Update) error {
	if update.Message == nil {
		return nil
	}

	if update.Message.Text == "/start" {
		a.handlers.HandleStart(ctx, a.sender, update)
		return nil
	}

	if update.Message.Text != "" {
		return a.handlers.HandleExpense(ctx, a.sender, update)
	}

	return nil
}

func main() {
	app, err := newApp()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize application: %v", err))
	}
	lambda.Start(app.handleRequest)
}
