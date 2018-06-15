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

var key contextKey

var level uint32

func Level(l uint32) { atomic.StoreUint32(&level, l) }

func ContextGetLogger(ctx context.Context) Logger {
	if v, ok := ctx.Value(key).(Logger); ok {
		return v
	}
	return Logger{}
}

func ContextWithLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, key, l)
}

type Logger struct {
	level  uint32
	prefix string
	output io.Writer
}

func (l Logger) V(level uint32) Logger {
	return Logger{
		level:  level,
		prefix: l.prefix,
		output: l.output,
	}
}

func (l Logger) Prefix(prefix string) Logger {
	return Logger{
		level:  l.level,
		prefix: "[" + prefix + "] ",
		output: l.output,
	}
}

func (l Logger) Print(msg string)                       { l.write(msg) }
func (l Logger) Printf(msg string, args ...interface{}) { l.write(msg, args...) }

func (l Logger) write(s string, args ...interface{}) {
	output := l.output
	if output == nil {
		output = os.Stderr
	}
	current := atomic.LoadUint32(&level)
	if l.level <= current {
		t := time.Now().UTC()
		msg := "[" + t.Format(time.RFC3339) + "] " + l.prefix + s + "\n"
		fmt.Fprintf(output, msg, args...)
	}
}
