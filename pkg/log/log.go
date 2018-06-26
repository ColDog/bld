package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

type contextKey int

var (
	key   contextKey
	level uint32
)

// Level will store a new global log level for the package.
func Level(l uint32) { atomic.StoreUint32(&level, l) }

// ContextGetLogger will return a logger stored in a context key.
func ContextGetLogger(ctx context.Context) Logger {
	if v, ok := ctx.Value(key).(Logger); ok {
		return v
	}
	return Logger{}
}

// ContextWithLogger returns a new context with a logger stored within it.
func ContextWithLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, key, l)
}

// Logger is a basic logger with V level verbosity.
type Logger struct {
	level  uint32
	prefix string
	output io.Writer
}

// Output returns a new logger with a changed output.
func (l Logger) Output(out io.Writer) Logger {
	return Logger{
		level:  l.level,
		prefix: l.prefix,
		output: out,
	}
}

// V will return a new Logger that will only log if the current level is greater
// than or equal to the provided level.
func (l Logger) V(level uint32) Logger {
	return Logger{
		level:  level,
		prefix: l.prefix,
		output: l.output,
	}
}

// Prefix sets a new prefix and returns a new Logger.
func (l Logger) Prefix(prefix string) Logger {
	return Logger{
		level:  l.level,
		prefix: "[" + prefix + "] ",
		output: l.output,
	}
}

// Print will print a new message.
func (l Logger) Print(msg string) { l.write(msg) }

// Printf will print and format.
func (l Logger) Printf(msg string, args ...interface{}) { l.write(msg, args...) }

func (l Logger) write(s string, args ...interface{}) {
	output := l.output
	if output == nil {
		output = os.Stdout
	}
	current := atomic.LoadUint32(&level)
	if l.level <= current {
		t := time.Now().UTC()
		msg := "[" + t.Format(time.RFC3339) + "] " + l.prefix + s + "\n"
		fmt.Fprintf(output, msg, args...)
	}
}
