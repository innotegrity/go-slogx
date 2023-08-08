package console_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fatih/color"
	"go.innotegrity.dev/slogx"
	"go.innotegrity.dev/slogx/handlers/console"
	"golang.org/x/exp/slog"
)

func TestMain(t *testing.T) {
	handlerOpts := console.Options{
		Level: slogx.LevelTrace,
		PartsOrder: []string{
			console.TimeKey,
			console.LevelKey,
			console.CallerKey,
			console.MessageKey,
			"{",
			fmt.Sprintf(console.AttrKey, "error"),
			fmt.Sprintf(console.AttrKey, "group.better"),
			console.AttrsKey,
			"}",
		},
		PartsFormatter: map[string]console.FormatterFn{
			fmt.Sprintf(console.AttrKey, "error"): errorFormatter,
			console.CallerKey:                     callerFormatter,
		},
	}
	handler := handlerOpts.NewHandler()
	logger := slog.New(handler)

	logger.Log(context.TODO(), slogx.LevelTrace, "this is a trace message")
	logger.Debug("this is a debug message")
	logger.Info("this is an info message")
	logger.Log(context.TODO(), slogx.LevelNotice, "this is a notice message")
	logger.Warn("this is a warning message")
	logger.Error("this is an error message")
	logger.Log(context.TODO(), slogx.LevelFatal, "this is a fatal message")
	logger.Log(context.TODO(), slogx.LevelPanic, "this is a panic message")

	logger.Info("this is an info message with attributes",
		slog.String("pie", "3.14"),
		slog.String("attr", "Value1"),
		slog.Int("attr2", 100),
		slog.Group("group",
			slog.String("group1Attr", "value")))
	logger.Error("this is an error message with attributes",
		slog.String("pie", "3.14"),
		slog.String("attr", "Value1"),
		slog.String("error", "this is the error message"),
		slog.Int("attr2", 100),
		slog.Group("group",
			slog.String("group1Attr", "value"),
			slog.String("better", "value"),
			slog.Int("a", 1),
		),
		slog.Group("user",
			slog.String("name", "josh"),
			slog.String("email", "josh@josh.com"),
		),
	)
}

func callerFormatter(ctx context.Context, key string, r slog.Record) string {
	frames := runtime.CallersFrames([]uintptr{r.PC})
	frame, _ := frames.Next()
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, frame.File); err == nil {
			frame.File = rel
		}
	}
	c := color.New(color.FgHiWhite)
	return c.Sprintf("%s:%d >", frame.File, frame.Line)
}

func errorFormatter(ctx context.Context, key string, r slog.Record) string {
	var errstr string
	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "error" {
			red := color.New(color.FgHiRed)
			key := red.Sprintf("error")
			errstr = fmt.Sprintf("%s=%s", key, attr.Value.String())
			return false
		}
		return true
	})
	return errstr
}
