package nlog

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

var logger *slog.Logger
var Leveler slog.Leveler

func Logger() *slog.Logger {
	if logger != nil {
		return logger
	}
	if Leveler == nil && os.Getenv("NSX_MODE") != "release" {
		Leveler = slog.LevelDebug
	} else {
		Leveler = slog.LevelInfo
	}
	logger = slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      Leveler,
		TimeFormat: time.Kitchen,
	}))
	return logger
}

func SetLevel(newLevel slog.Leveler) {
	Leveler = newLevel
}

func SetWriter(w io.Writer) {
	logger = slog.New(tint.NewHandler(w, &tint.Options{
		Level:      Leveler,
		TimeFormat: time.Kitchen,
	}))
}

