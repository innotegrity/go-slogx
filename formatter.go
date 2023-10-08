package slogx

import "context"

// formatterOptionsContextKey is used to store a formatter's options in a standard Go context object.
type formatterOptionsContextKey struct {
	name string
}

// ContextWithFormatterOptions copies the given context and returns a new context with the given formatter options
// stored in it with the given name.
func ContextWithFormatterOptions(ctx context.Context, opts any, name string) context.Context {
	return context.WithValue(ctx, formatterOptionsContextKey{name: name}, opts)
}

// FormatterOptionsFromContext retrieves the formatter options stored in the given context with the given name, if
// it exists.
//
// If the handler options cannot be found, nil is returned.
func FormatterOptionsFromContext(ctx context.Context, name string) any {
	if v := ctx.Value(formatterOptionsContextKey{name: name}); v != nil {
		return v
	}
	return nil
}
