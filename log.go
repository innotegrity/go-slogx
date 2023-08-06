package slog

/*
type tblogKey struct{}

type Logger struct {
	internal *slog.Logger
}

func New(handler slog.Handler) *Logger {
	return &Logger{
		internal: slog.New(handler),
	}
}

func (l Logger) GetLogger() *slog.Logger {
	return l.internal
}

func (l Logger) WithLogger(logger *slog.Logger) *Logger {
	l.internal = logger
	return &l
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(tblogKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func (l Logger) AddToContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, tblogKey{}, l)
}

func Err(msg ...string) slog.Attr {
	return slog.Group(E
		Key:   ErrorKey,
		Value: slog.StringValue(msg),
	}
}
*/
