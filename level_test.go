package slogx_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"go.innotegrity.dev/slogx"
	"golang.org/x/exp/slog"
)

// TODO: implement testing and benchmarks

func TestLevel(t *testing.T) {
	for i := -20; i < 20; i++ {
		t.Logf("Level as string: %s", slogx.Level(i))
		t.Logf("Level as short string: %s", slogx.Level(i).ShortString())
	}

	t.Logf("Trace Level as string: %s", slogx.LevelTrace)
	b, _ := slogx.LevelTrace.MarshalJSON()
	t.Logf("Trace Level as JSON: %s", string(b))
	b, _ = slogx.LevelTrace.MarshalText()
	t.Logf("Trace Level as text: %s", string(b))
	var l slogx.Level
	if err := json.Unmarshal([]byte(`"trace"`), &l); err != nil {
		t.Logf("error: %s", err.Error())
	} else {
		t.Logf("unmarshaled: %s %d", l, l)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(-99)}))
	logger.Log(context.TODO(), l.Level(), "this is a message")
}
