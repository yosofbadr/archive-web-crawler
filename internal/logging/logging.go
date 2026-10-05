package logging

import (
	"io"
	"log"
	"os"
)

// Logger is a thin wrapper to allow swapping implementations.
// It deliberately mirrors log.Logger subset needed by the app.
type Logger interface {
	Printf(format string, v ...any)
	Println(v ...any)
}

type stdLogger struct{ *log.Logger }

// New creates a production-friendly logger writing to stdout.
func New() Logger {
	return &stdLogger{Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)}
}

// NewWithWriter creates a logger writing to a custom writer (useful for tests).
func NewWithWriter(w io.Writer) Logger {
	return &stdLogger{Logger: log.New(w, "", log.LstdFlags|log.Lmicroseconds)}
}
