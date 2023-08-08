package main

import (
	"context"

	slogslack "github.com/samber/slog-slack"
	slogmulti "go.innotegrity.dev/slogx/handlers/multiwriter"
	"golang.org/x/exp/slog"
)

func recordMatchRegion(region string) func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
		ok := false

		r.Attrs(func(attr slog.Attr) bool {
			if attr.Key == "region" && attr.Value.Kind() == slog.KindString && attr.Value.String() == region {
				ok = true
				return false
			}

			return true
		})

		return ok
	}
}

func main() {
	slackChannelUS := slogslack.Option{Level: slog.LevelError, WebhookURL: "xxx", Channel: "supervision-us"}.NewSlackHandler()
	slackChannelEU := slogslack.Option{Level: slog.LevelError, WebhookURL: "xxx", Channel: "supervision-eu"}.NewSlackHandler()
	slackChannelAPAC := slogslack.Option{Level: slog.LevelError, WebhookURL: "xxx", Channel: "supervision-apac"}.NewSlackHandler()

	logger := slog.New(
		slogmulti.Router().
			Add(slackChannelUS, recordMatchRegion("us")).
			Add(slackChannelEU, recordMatchRegion("eu")).
			Add(slackChannelAPAC, recordMatchRegion("apac")).
			Handler(),
	)

	logger.
		With("region", "us").
		With("pool", "us-west-1").
		Error("Server desynchronized")
}
