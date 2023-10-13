package handler_test

// TODO: implement testing and benchmarks

import (
	"errors"
	"os"
	"testing"
	"time"

	"log/slog"

	"github.com/mattn/go-colorable"
	"go.innotegrity.dev/errorx"
	"go.innotegrity.dev/slogx"
	"go.innotegrity.dev/slogx/handler"
)

func TestMultiHandler(t *testing.T) {
	writer := colorable.NewColorable(os.Stdout)
	consoleHandler := handler.NewConsoleHandler(handler.ConsoleHandlerOptions{
		Level:           slogx.NewLevelVar(slogx.LevelTrace),
		RecordFormatter: nil,
		Writer:          writer,
	})
	jsonHandler := handler.NewJSONHandler(handler.JSONHandlerOptions{
		Level:           slogx.NewLevelVar(slogx.LevelTrace),
		RecordFormatter: nil,
		Writer:          os.Stdout,
	})
	logger := slogx.Wrap(
		slog.New(
			handler.NewMultiHandler(
				handler.MultiHandlerOptions{},
				consoleHandler,
				jsonHandler,
			),
		),
	)

	logger.Trace("this is a trace message", slog.Duration("duration", time.Second*5), slog.Float64("pi", 3.141569))
	logger.Debug("this is a debug message")
	logger.Info("this is an info message")
	logger.Notice("this is a notice message")
	logger = slogx.Wrap(logger.With(slog.String("root_key", "1")).WithGroup("group1").With(slog.String("k1", "v1")).WithGroup("nested").With(slog.String("logger_name", "frodo")))
	logger.Warn("this is a warning message")
	logger.Error("this is an error message")
	logger.Fatal("this is a fatal message", slog.Int("exit_code", 10))
	logger.Panic("this is a panic message")

	logger.Info("this is an info message with attributes",
		slog.Float64("pie", 3.14),
		slog.String("attr", "Value1"),
		slog.Int("attr2", 100),
		slog.Duration("took", time.Second*5),
		slog.Time("now", time.Now()),
		slog.Group("group",
			slog.String("group1Attr", "value")),
		slog.String("logger_name", "precious"))
	logger.Error("this is an error message with attributes",
		slog.Float64("pie", 3.141579),
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
		slog.Any("admin", User{
			Username: "admin",
			Password: "admin123",
			Addresses: []Address{
				{
					Street:     "1234 Acme Way",
					City:       "New York",
					PostalCode: "12345",
					Country:    "United States",
				},
				{
					Street:     "555 Sunset Blvd",
					City:       "Hollywood",
					PostalCode: "90028",
					Country:    "United States",
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
