package logutils

import (
	"fmt"
	"log/slog"
	"strings"
)

func LogLevelFromString(s string) (slog.Level, error) {

	switch strings.ToLower(s) {
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	case "debug":
		return slog.LevelDebug, nil
	}

	return slog.LevelInfo, fmt.Errorf("invalid log level")
}
