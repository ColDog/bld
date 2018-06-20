package log

import (
	"context"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	Level(1)

	Logger{}.Prefix("test").V(1).Print("hello")
	Logger{}.Prefix("test").V(2).Printf("hello %s", "1")
	Logger{}.Prefix("test").Output(os.Stdout).V(3).Print("hello")
}

func TestLogContext(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithLogger(ctx, Logger{}.Prefix("test"))
	ContextGetLogger(ctx).Print("hi")
	ContextGetLogger(context.Background()).Print("hi")
}
