package handlers_test

// TODO: implement testing and benchmarks

import (
	"errors"
	"testing"
	"time"

	"go.innotegrity.dev/errorx"
	"go.innotegrity.dev/slogx"
	"go.innotegrity.dev/slogx/formatters"
	"go.innotegrity.dev/slogx/handlers"
	"golang.org/x/exp/slog"
)

func TestSlack1(t *testing.T) {
	slackFormatterOptions := formatters.DefaultSlackMessageFormatterOptions()
	slackFormatterOptions.ApplicationName = "StrongBox One"
	slackFormatterOptions.ApplicationIconURL = "https://cdn.pixabay.com/photo/2012/04/12/13/49/safe-30110_960_720.png"
	slackFormatterOptions.IncludeSource = true
	slackFormatter := formatters.NewSlackMessageFormatter(slackFormatterOptions)
	slackHandler, _ := handlers.NewSlackHandler(handlers.SlackHandlerOptions{
		EnableAsync:     true,
		Level:           slogx.LevelTrace,
		RecordFormatter: slackFormatter,
		WebhookURL:      "https://hooks.slack.com/services/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",
	})
	logger := slogx.Wrap(slog.New(slackHandler))
	defer logger.Shutdown(true)

	logger.Trace("this is a trace message")
	logger.Debug("this is a debug message")
	logger.Info("this is an info message")
	logger.Notice("this is a notice message")
	logger.Warn("this is a warning message")
	logger = slogx.Wrap(logger.With(slog.String("root_key", "1")).WithGroup("group1").With(slog.String("k1", "v1")).WithGroup("nested").With(slog.String("logger_name", "frodo")))
	logger.Error("this is an error message")
	logger.Fatal("this is a fatal message")
	logger.Panic("this is a panic message")

	logger.Info("this is an info message with attributes",
		slog.Float64("pie", 3.14),
		slog.String("attr", "Value1"),
		slog.Int("attr2", 100),
		slog.Duration("took", time.Second*5),
		slog.Time("now", time.Now()),
		slog.Group("group",
			slog.String("group1Attr", "value")))
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
