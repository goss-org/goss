package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/goss-org/goss/util"
)

// parseLogLevel resolves a --log-level value to a slog.Level. The five names are
// the ones the message-prefix filter this replaces accepted, matched the same
// case-insensitive way, so no spelling that works today stops working.
func parseLogLevel(level string) (slog.Level, error) {
	switch strings.ToUpper(level) {
	case "TRACE":
		return util.LevelTrace, nil
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	}
	return 0, fmt.Errorf("unsupported log level: %s", level)
}

// newCLIHandler builds the handler the goss CLI logs through. Production always
// passes os.Stderr; the writer is a parameter so that a test can assert on the
// bytes a record actually renders to.
//
// The CLI owns two policies here. It renders util.LevelTrace as TRACE, using the
// same exported hook an embedder would install, and it emits timestamps in UTC,
// as the log output of the paths that set a level does today.
func newCLIHandler(w io.Writer, level slog.Leveler) slog.Handler {
	return &utcHandler{
		handler: slog.NewTextHandler(w, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: util.ReplaceTraceLevel,
		}),
	}
}

// utcHandler normalises each record's timestamp to UTC and delegates the rest.
//
// This is a handler wrapper rather than a ReplaceAttr function because
// ReplaceAttr cannot tell the built-in time field from a user attribute that
// happens to be named "time": both arrive as an attribute with that key. A
// wrapper sees slog.Record.Time itself, so it can convert one and leave the
// other alone.
type utcHandler struct {
	handler slog.Handler
}

func (h *utcHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *utcHandler) Handle(ctx context.Context, record slog.Record) error {
	// A zero time means "no timestamp" to a handler, so leave it alone.
	if !record.Time.IsZero() {
		record.Time = record.Time.UTC()
	}
	return h.handler.Handle(ctx, record)
}

func (h *utcHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &utcHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *utcHandler) WithGroup(name string) slog.Handler {
	return &utcHandler{handler: h.handler.WithGroup(name)}
}
