package log

import "testing"

func TestLog(t *testing.T) {
	Level(1)

	Logger{}.Prefix("test").V(1).Print("hello")
	Logger{}.Prefix("test").V(2).Printf("hello %s", "1")
}
