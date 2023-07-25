package handler

import (
	"context"
	"io"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/fluxx1on/thumbnails_microservice/libs/logger/attrs"
	"github.com/spf13/cast"
	"golang.org/x/exp/slog"
)

type ColorfulHandler struct {
	// implements base struct
	slog.Handler

	logger     *log.Logger
	logJournal *log.Logger
	attrs      []slog.Attr
}

func NewColorfulHandler(out, journal io.Writer, opts *slog.HandlerOptions) *ColorfulHandler {
	h := &ColorfulHandler{
		Handler:    slog.NewTextHandler(out, opts),
		logger:     log.New(out, "", 0),
		logJournal: log.New(journal, "", 0),
	}

	return h
}

func (h *ColorfulHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.HiBlackString(level)
	case slog.LevelInfo:
		level = color.GreenString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	textTrace := &strings.Builder{}
	textTrace.Grow(r.NumAttrs())

	decorator := func(a slog.Attr) {
		var b slog.Attr
		if err, is := a.Value.Any().(error); is {
			b = attrs.Err(err)
		} else {
			b = attrs.Any(a.Value)
		}

		textTrace.WriteString(cast.ToString(b.Key) + " ")
		textTrace.WriteString(cast.ToString(b.Value.Any()))
	}

	r.Attrs(func(a slog.Attr) bool {
		decorator(a)

		return true
	})

	for _, a := range h.attrs {
		decorator(a)
	}

	timeStr := r.Time.Format("[15:04:05.000]")
	msg := color.CyanString(r.Message)
	additionalInfo := color.WhiteString(textTrace.String())

	h.logJournal.Println(
		timeStr,
		level,
		r.Message,
		textTrace.String(),
	)

	h.logger.Println(
		timeStr,
		level,
		msg,
		color.WhiteString(string(additionalInfo)),
	)

	return nil
}
