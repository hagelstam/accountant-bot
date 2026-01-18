package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	TelegramBotToken      string
	GoogleCredentialsJSON string
	GoogleSpreadsheetID   string
	LoggingLevel          slog.Level
}

func LoadConfig() (*Config, error) {
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if telegramToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	googleCreds := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if googleCreds == "" {
		return nil, fmt.Errorf("GOOGLE_CREDENTIALS_JSON environment variable is required")
	}

	spreadsheetID := os.Getenv("GOOGLE_SPREADSHEET_ID")
	if spreadsheetID == "" {
		return nil, fmt.Errorf("GOOGLE_SPREADSHEET_ID environment variable is required")
	}

	loggingLevel := getLoggingLevel(os.Getenv("LOGGING_LEVEL"))

	return &Config{
		TelegramBotToken:      telegramToken,
		GoogleCredentialsJSON: googleCreds,
		GoogleSpreadsheetID:   spreadsheetID,
		LoggingLevel:          loggingLevel,
	}, nil
}

func getLoggingLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
