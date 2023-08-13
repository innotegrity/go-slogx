package utils

import (
	"context"
	"fmt"
	"sort"

	"golang.org/x/exp/slog"
)

// FlattenRecordAttrs resolves all attributes and group attributes and "flattens" everything out.
func FlattenRecordAttrs(ctx context.Context, r slog.Record, sortKeys bool) (
	[]string, map[string]slog.Value) {

	attrs := []slog.Attr{}
	r.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})

	keys, flattenedAttrs := flattenAttrs(ctx, attrs, "")
	if sortKeys {
		sort.Strings(keys)
	}
	return keys, flattenedAttrs
}

// flattenAttrs resolves all group attributes and "flattens" them out.
func flattenAttrs(ctx context.Context, attrs []slog.Attr, keyPrefix string) (
	[]string, map[string]slog.Value) {

	flattened := map[string]slog.Value{}
	attrKeys := []string{}
	for _, attr := range attrs {
		attrKey := attr.Key
		if keyPrefix != "" {
			attrKey = fmt.Sprintf("%s.%s", keyPrefix, attrKey)
		}

		// flatten all group attributes
		if attr.Value.Kind() == slog.KindGroup {
			groupKeys, groupAttr := flattenAttrs(ctx, attr.Value.Group(), attrKey)
			attrKeys = append(attrKeys, groupKeys...)
			for k, v := range groupAttr {
				flattened[k] = v
			}
			continue
		}

		// add keys and values
		attrKeys = append(attrKeys, attrKey)
		flattened[attrKey] = attr.Value
	}
	return attrKeys, flattened
}
