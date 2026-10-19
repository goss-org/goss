package util

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

// LevelTrace is the level goss emits its most detailed diagnostics at. It sits
// below slog.LevelDebug, so a logger configured at DEBUG or above suppresses
// them.
//
// It is a constant rather than a variable so that no caller can mutate goss's
// level convention, and so that concurrent readers need no synchronisation.
const LevelTrace slog.Level = -8

// levelTraceLabel is how LevelTrace renders. Standard slog handlers have no
// name for a level this far below DEBUG and would render it as "DEBUG-4".
const levelTraceLabel = "TRACE"

// discardLogger is the logger LoggerOrDiscard substitutes for a nil one. goss
// deliberately has no slog.Default() fallback: a library that writes records
// somewhere the embedder did not choose is the defect this replaces.
var discardLogger = slog.New(slog.DiscardHandler)

// ReplaceTraceLevel renders LevelTrace as TRACE. It has the exact signature of
// slog.HandlerOptions.ReplaceAttr, so an embedder can install it on their own
// handler and have goss's trace records render the way the goss CLI renders
// them:
//
//	slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
//		Level:       util.LevelTrace,
//		ReplaceAttr: util.ReplaceTraceLevel,
//	})
//
// It rewrites only the built-in level attribute, and only when that attribute
// holds LevelTrace, so it composes with a caller's own ReplaceAttr function.
func ReplaceTraceLevel(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) != 0 || attr.Key != slog.LevelKey {
		return attr
	}
	if level, ok := attr.Value.Any().(slog.Level); ok && level == LevelTrace {
		attr.Value = slog.StringValue(levelTraceLabel)
	}
	return attr
}

// LoggerOrDiscard returns l, or a logger that discards every record when l is
// nil. Every derivation of a logger from a Config or an OutputConfig goes
// through this guard: util.Config is documented as being usable as a composite
// literal, so its Logger field is nil in ordinary embedded use, and a nil
// *slog.Logger panics on every method that matters.
func LoggerOrDiscard(l *slog.Logger) *slog.Logger {
	if l != nil {
		return l
	}
	return discardLogger
}

// Trace emits msg and args at LevelTrace through l. A nil l discards the
// record. The record's source position is the caller's, not this function's.
func Trace(l *slog.Logger, msg string, args ...any) {
	trace(context.Background(), l, msg, args...)
}

// TraceContext is Trace with an explicit context, which is passed to the
// handler. Note that goss honours nothing else on that context yet, and that
// cancelling it does not by itself suppress a record: whether to do that is the
// handler's choice.
func TraceContext(ctx context.Context, l *slog.Logger, msg string, args ...any) {
	trace(ctx, l, msg, args...)
}

// trace builds the record by hand rather than calling Logger.Log, so that the
// program counter it carries is the exported wrapper's caller.
func trace(ctx context.Context, l *slog.Logger, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	logger := LoggerOrDiscard(l)
	if !logger.Enabled(ctx, LevelTrace) {
		return
	}

	// Skip runtime.Callers, this function and the exported wrapper.
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])

	record := slog.NewRecord(time.Now(), LevelTrace, msg, pcs[0])
	record.Add(args...)
	_ = logger.Handler().Handle(ctx, record)
}
