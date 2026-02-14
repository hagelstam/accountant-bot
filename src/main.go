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

type application struct {
	bot      *bot.Bot
	handlers *BotHandlers
	logger   *slog.Logger
}

func newApplication() (*application, error) {
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

	return &application{
		bot:      telegramBot,
		handlers: handlers,
		logger:   logger,
	}, nil
}

func (app *application) handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var update models.Update
	if err := json.Unmarshal([]byte(request.Body), &update); err != nil {
		app.logger.Error("failed to unmarshal update", slog.String("error", err.Error()))
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Invalid JSON"}`,
		}, nil
	}

	if err := app.processUpdate(ctx, &update); err != nil {
		app.logger.Error("failed to process update", slog.String("error", err.Error()))
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Internal server error"}`,
		}, nil
	}

	app.logger.Info("successfully processed update", slog.Int64("update_id", update.ID))

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"status": "ok"}`,
	}, nil
}

func (app *application) processUpdate(ctx context.Context, update *models.Update) error {
	if update.Message == nil {
		return nil
	}

	// Handle /start command
	if update.Message.Text == "/start" {
		app.handlers.HandleStart(ctx, app.bot, update)
		return nil
	}

	// Handle expenses
	if update.Message.Text != "" {
		app.handlers.HandleExpense(ctx, app.bot, update)
		return nil
	}

	return nil
}

func main() {
	app, err := newApplication()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize application: %v", err))
	}

	lambda.Start(app.handleRequest)
}
