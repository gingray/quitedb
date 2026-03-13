package app

import (
	"log/slog"
	"os"
)

func (a *App) WithLogger() error {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.Time(slog.TimeKey, t.UTC())
			}
			return a
		},
	})

	a.Logger = slog.New(handler)
	return nil
}
