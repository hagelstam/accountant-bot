package main

import (
	"log/slog"
	"testing"
)

func TestGetLoggingLevel(t *testing.T) {
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

			got := getLoggingLevel(tt.level)
			if got != tt.want {
				t.Errorf("getLoggingLevel(%q) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}
