package clog

import (
	"log/slog"
	"os"

	"github.com/tgweber/mlmetaretrieve/internal/config"
)

func SetupLogger(config config.Config) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
