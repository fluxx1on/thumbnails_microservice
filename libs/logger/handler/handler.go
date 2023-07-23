package handler

import (
	"context"
	"io"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cast"
	"golang.org/x/exp/slog"
)

type ColorfulHandler struct {
	// implements base struct
	slog.Handler

	// opts   *slog.HandlerOptions
	logger *log.Logger
	attrs  []slog.Attr
}

func NewColorfulHandler(out io.Writer) *ColorfulHandler {
	h := &ColorfulHandler{
		Handler: slog.NewTextHandler(out, &slog.HandlerOptions{}),
		logger:  log.New(out, "", 0),
	}

	return h
}

func (h *ColorfulHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.CyanString(level)
	case slog.LevelInfo:
		level = color.GreenString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	textTrace := &strings.Builder{}
	textTrace.Grow(r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		textTrace.WriteString(cast.ToString(a.Key) + " ")
		textTrace.WriteString(cast.ToString(a.Value.Any()))

		return true
	})

	for _, a := range h.attrs {
		textTrace.WriteString(cast.ToString(a.Key) + " ")
		textTrace.WriteString(cast.ToString(a.Value.Any()))
	}

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := color.CyanString(r.Message)
	additionalInfo := color.WhiteString(textTrace.String())

	h.logger.Println(
		timeStr,
		level,
		msg,
		color.WhiteString(string(additionalInfo)),
	)

	return nil
}
