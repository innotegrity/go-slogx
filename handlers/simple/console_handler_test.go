package simple_test

// TODO: implement testing and benchmarks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"go.innotegrity.dev/errorx"
	"go.innotegrity.dev/runtimex"
	"go.innotegrity.dev/slogx"
	"go.innotegrity.dev/slogx/handlers/simple"
	"go.innotegrity.dev/slogx/middleware"
	"golang.org/x/exp/slog"
)

func TestMain(t *testing.T) {
	writer := colorable.NewColorable(os.Stdout)
	handler := simple.NewConsoleHandler(simple.ConsoleHandlerOptions{
		AttrFormatter: formatAttr,
		//ContinueOnError: false,
		EnableColor:    true,
		Level:          slogx.LevelTrace,
		LevelFormatter: formatLevel,
		//IgnoreAttrs:      []string{},
		//MessageFormatter: nil,
		PartOrder: []string{
			simple.TimePart,
			simple.LevelPart,
			simple.SourcePart,
			">",
			simple.MessagePart,
			"{",
			fmt.Sprintf(simple.AttrPart, "error"),
			simple.AllAttrsAlphaPart,
			"}",
		},
		//PartSeparator: " ",
		SourceFormatter: formatSource,
		SpecificAttrFormatter: map[string]middleware.FormatAttrFn{
			"error": formatError,
		},
		TimeFormatter: formatTime,
		Writer:        writer,
	})
	logger := slogx.Wrap(slog.New(handler))

	logger.Trace("this is a trace message")
	logger.Debug("this is a debug message")
	logger.Info("this is an info message")
	logger.Notice("this is a notice message")
	logger.Warn("this is a warning message")
	logger.Error("this is an error message")
	logger.Fatal("this is a fatal message")
	logger.Panic("this is a panic message")

	logger.Info("this is an info message with attributes",
		slog.String("pie", "3.14"),
		slog.String("attr", "Value1"),
		slog.Int("attr2", 100),
		slog.Group("group",
			slog.String("group1Attr", "value")))
	logger.Error("this is an error message with attributes",
		slog.String("pie", "3.14"),
		slog.String("attr", "Value1"),
		slogx.Err("error", errors.New("this is the error message")),
		slogx.ErrX("extended_error", &ErrTest{
			Value1: "important",
			Value2: 1234,
			Err:    errors.New("some error"),
			NestedErr: []errorx.Error{
				&ErrTest{
					Value1: "not so important",
					Value2: 3345,
					Err:    errors.New("some other error"),
				},
			},
		}),
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

func formatAttr(ctx context.Context, attrKey string, attrValue slog.Value) (string, slog.Value, error) {
	if v := ctx.Value(simple.ConsoleHandlerOptionsContext{}); v != nil {
		if opts, ok := v.(*simple.ConsoleHandlerOptions); ok && opts.EnableColor {
			c := color.New(color.FgHiBlue)
			return c.Sprint(attrKey), attrValue, nil
		}
	}
	return attrKey, attrValue, nil
}

func formatError(ctx context.Context, attrKey string, attrValue slog.Value) (string, slog.Value, error) {
	if v := ctx.Value(simple.ConsoleHandlerOptionsContext{}); v != nil {
		if opts, ok := v.(*simple.ConsoleHandlerOptions); ok && opts.EnableColor {
			c := color.New(color.FgHiRed)
			return c.Sprint(attrKey), attrValue, nil
		}
	}
	return attrKey, attrValue, nil
}

func formatLevel(ctx context.Context, level slogx.Level) (slog.Value, error) {
	var c *color.Color
	if v := ctx.Value(simple.ConsoleHandlerOptionsContext{}); v != nil {
		if opts, ok := v.(*simple.ConsoleHandlerOptions); ok && opts.EnableColor {
			switch level {
			case slogx.LevelTrace:
				c = color.New(color.FgHiMagenta)
			case slogx.LevelDebug:
				c = color.New(color.FgBlue)
			case slogx.LevelInfo:
				c = color.New(color.FgHiGreen)
			case slogx.LevelNotice:
				c = color.New(color.FgHiYellow)
			case slogx.LevelWarn:
				c = color.New(color.FgYellow)
			case slogx.LevelError:
				c = color.New(color.FgHiRed)
			case slogx.LevelFatal:
				c = color.New(color.FgRed, color.BgWhite)
			case slogx.LevelPanic:
				c = color.New(color.FgRed, color.BgHiYellow)
			default:
				c = nil
			}
		}
	}
	if c != nil {
		return slog.StringValue(c.Sprintf("%s", level.ShortString())), nil
	}
	return slog.StringValue(level.ShortString()), nil
}

func formatSource(ctx context.Context, pc uintptr) (slog.Value, error) {
	f := runtimex.FrameFromPC(pc)
	if v := ctx.Value(simple.ConsoleHandlerOptionsContext{}); v != nil {
		if opts, ok := v.(*simple.ConsoleHandlerOptions); ok && opts.EnableColor {
			c := color.New(color.FgHiWhite)
			return slog.StringValue(c.Sprintf("%s", f)), nil
		}
	}
	return slog.StringValue(fmt.Sprintf("%s", f)), nil
}

func formatTime(ctx context.Context, t time.Time) (slog.Value, error) {
	return slog.StringValue(t.Format("15:04:05PM")), nil
}

const (
	ErrTestCode = 1751
)

type ErrTest struct {
	Err       error
	Value1    string
	Value2    int
	NestedErr []errorx.Error
}

// InternalError returns the internal standard error object if there is one or nil if none is set.
func (e *ErrTest) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *ErrTest) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("an error has occurred: %s", e.Err.Error())
	}
	return "an error has occurred"
}

// Code returns the corresponding error code.
func (e *ErrTest) Code() int {
	return ErrTestCode
}

func (e *ErrTest) Attrs() map[string]any {
	return map[string]any{
		"value1": e.Value1,
		"value2": e.Value2,
	}
}

func (e *ErrTest) NestedErrors() []errorx.Error {
	return e.NestedErr
}
