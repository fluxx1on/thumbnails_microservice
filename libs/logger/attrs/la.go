package attrs

import (
	"strings"

	"github.com/spf13/cast"
	"golang.org/x/exp/slog"
)

func Err(errs ...error) slog.Attr {
	var (
		key    string = "(__error__)"
		strerr []string
	)

	if len(errs) > 1 {
		key = "(__errors__)"
	}

	strerr = make([]string, len(errs))
	for i, err := range errs {
		strerr[i] = err.Error()
	}

	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(strings.Join(strerr, " ; ")),
	}
}

func Any(values ...any) slog.Attr {
	var (
		key    string = "(__param__)"
		params []string
	)

	if len(values) > 1 {
		key = "(__params__)"
	}

	params = make([]string, len(values))
	for i, value := range values {
		params[i] = cast.ToString(value)
	}

	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(strings.Join(params, " ; ")),
	}
}
