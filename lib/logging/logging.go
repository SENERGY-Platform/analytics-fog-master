package logging

import (
	"io"
	"log/slog"
)


var Logger *slog.Logger

func InitLogger(outputWriter io.Writer,debug bool) (err error) {
	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}
	Logger = slog.New(slog.NewJSONHandler(outputWriter, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(Logger)
	return
}

