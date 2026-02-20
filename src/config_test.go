package main

import (
	"log/slog"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	envKeys := []string{"TELEGRAM_BOT_TOKEN", "GOOGLE_CREDENTIALS_JSON", "GOOGLE_SPREADSHEET_ID", "LOG_LEVEL"}

	tests := []struct {
		name    string
		envVars map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name: "valid config with defaults",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN":      "test-token",
				"GOOGLE_CREDENTIALS_JSON": `{"type":"service_account"}`,
				"GOOGLE_SPREADSHEET_ID":   "sheet-123",
			},
			want: &Config{
				TelegramBotToken:      "test-token",
				GoogleCredentialsJSON: `{"type":"service_account"}`,
				GoogleSpreadsheetID:   "sheet-123",
				LogLevel:              slog.LevelInfo,
			},
		},
		{
			name: "valid config with log level",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN":      "test-token",
				"GOOGLE_CREDENTIALS_JSON": `{"type":"service_account"}`,
				"GOOGLE_SPREADSHEET_ID":   "sheet-123",
				"LOG_LEVEL":               "DEBUG",
			},
			want: &Config{
				TelegramBotToken:      "test-token",
				GoogleCredentialsJSON: `{"type":"service_account"}`,
				GoogleSpreadsheetID:   "sheet-123",
				LogLevel:              slog.LevelDebug,
			},
		},
		{
			name: "missing telegram token",
			envVars: map[string]string{
				"GOOGLE_CREDENTIALS_JSON": `{"type":"service_account"}`,
				"GOOGLE_SPREADSHEET_ID":   "sheet-123",
			},
			wantErr: true,
		},
		{
			name: "missing google credentials",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN":    "test-token",
				"GOOGLE_SPREADSHEET_ID": "sheet-123",
			},
			wantErr: true,
		},
		{
			name: "missing spreadsheet id",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN":      "test-token",
				"GOOGLE_CREDENTIALS_JSON": `{"type":"service_account"}`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range envKeys {
				t.Setenv(key, "")
			}
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			config, err := LoadConfig()

			if tt.wantErr {
				if err == nil {
					t.Error("LoadConfig() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("LoadConfig() unexpected error: %v", err)
			}
			if *config != *tt.want {
				t.Errorf("LoadConfig() = %+v, want %+v", *config, *tt.want)
			}
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level string
		want  slog.Level
	}{
		{
			name:  "debug uppercase",
			level: "DEBUG",
			want:  slog.LevelDebug,
		},
		{
			name:  "debug lowercase",
			level: "debug",
			want:  slog.LevelDebug,
		},
		{
			name:  "info uppercase",
			level: "INFO",
			want:  slog.LevelInfo,
		},
		{
			name:  "info lowercase",
			level: "info",
			want:  slog.LevelInfo,
		},
		{
			name:  "warn uppercase",
			level: "WARN",
			want:  slog.LevelWarn,
		},
		{
			name:  "warn lowercase",
			level: "warn",
			want:  slog.LevelWarn,
		},
		{
			name:  "error uppercase",
			level: "ERROR",
			want:  slog.LevelError,
		},
		{
			name:  "error lowercase",
			level: "error",
			want:  slog.LevelError,
		},
		{
			name:  "empty string defaults to info",
			level: "",
			want:  slog.LevelInfo,
		},
		{
			name:  "invalid string defaults to info",
			level: "invalid",
			want:  slog.LevelInfo,
		},
		{
			name:  "mixed case handled correctly",
			level: "DeBuG",
			want:  slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getLogLevel(tt.level)
			if got != tt.want {
				t.Errorf("getLogLevel(%q) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}
