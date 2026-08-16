package system

import (
	"bytes"
	"log/slog"

	"github.com/goss-org/goss/util"
)

// Streams a subject command's output can arrive on.
const (
	streamStdout = "stdout"
	streamStderr = "stderr"
)

// logCommandOutput emits one record per line of a command's output.
//
// The line splitting is deliberately unchanged, trailing empty element and all:
// output ending in a newline produced an empty final record before, and a
// consumer counting records would notice it going away.
func logCommandOutput(logger *slog.Logger, b []byte, resourceID, command, stream string) {
	if len(b) == 0 {
		return
	}

	l := util.LoggerOrDiscard(logger)
	for _, line := range bytes.Split(b, []byte("\n")) {
		l.Debug("command output",
			"resource_id", resourceID,
			"command", command,
			"stream", stream,
			"output", string(line))
	}
}
