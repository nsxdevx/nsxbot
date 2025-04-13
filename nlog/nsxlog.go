package nlog

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

var logger *slog.Logger

func Logger() *slog.Logger {
	if logger != nil {
		return logger
	}
	logger = slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
	}))
	return logger
}
