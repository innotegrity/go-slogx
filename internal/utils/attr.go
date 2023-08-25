package utils

import (
	"fmt"
	"sort"

	"go.innotegrity.dev/generic"
	"golang.org/x/exp/slog"
)

// ConsolidateAttrs combines the given attributes with attributes from the record, mapping the record attributes under
// the group, if not empty.
//
// Attribute values are resolved during the consolidation and duplicate attributes are removed from the returned slice
// and any nested groups. If an attribute is specified more than once, the last one specified is used.
func ConsolidateAttrs(attrs []slog.Attr, group string, record slog.Record) []slog.Attr {
	result := attrs

	if group == "" {
		record.Attrs(func(attr slog.Attr) bool {
			result = append(result, attr)
			return true
		})
	} else {
		groupAttrs := []any{}
		record.Attrs(func(attr slog.Attr) bool {
			groupAttrs = append(groupAttrs, attr)
			return true
		})
		result = append(result, slog.Group(group, groupAttrs...))
	}
	return UniqAttrs(result)
}

// FlattenAttrs takes the given slice of attributes and recursively "flattens" groups changing the attribute keys to
// GROUP.KEY (or GROUP.GROUP.KEY in the case of nested groups).
func FlattenAttrs(attrs []slog.Attr) []slog.Attr {
	result := []slog.Attr{}
	for _, attr := range attrs {
		if attr.Value.Kind() == slog.KindGroup {
			groupAttrs := FlattenAttrs(attr.Value.Group())
			for _, groupAttr := range groupAttrs {
				result = append(result, slog.Attr{Key: fmt.Sprintf("%s.%s", attr.Key, groupAttr.Key), Value: groupAttr.Value})
			}
		} else {
			result = append(result, attr)
		}
	}
	return result
}

// SortAttrs sorts the given attributes and returns a slice sorted by attribute key.
//
// Any nested attribute groups are sorted by attribute key as well.
func SortAttrs(attrs []slog.Attr) []slog.Attr {
	attrMap := map[string]slog.Value{}
	keySet := generic.NewSet[string]()
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
		keySet.Add(attr.Key)
	}
	keys := keySet.Members()
	sort.Strings(keys)

	result := []slog.Attr{}
	for _, key := range keys {
		if attrMap[key].Kind() == slog.KindGroup {
			result = append(result, slog.Group(key, generic.AnySlice(SortAttrs(attrMap[key].Group()))...))
		} else {
			result = append(result, slog.Attr{Key: key, Value: attrMap[key]})
		}
	}
	return result
}

// ToAttrMap converts the given attribute slice to a map of string/values.
//
// This function does not recursively convert groups. Use FlattenAttrs to flatten the attribute list first.
func ToAttrMap(attrs []slog.Attr) map[string]slog.Value {
	result := map[string]slog.Value{}
	for _, attr := range attrs {
		result[attr.Key] = attr.Value
	}
	return result
}

// UniqAttrs removes duplicate attributes from the slice and any nested groups, resolving attribute values along
// the way.
//
// If an attribute is duplicated, the last duplicate entry is used in the resulting slice.
func UniqAttrs(attrs []slog.Attr) []slog.Attr {
	seen := generic.NewSet[string]()
	result := []slog.Attr{}

	for i := len(attrs) - 1; i >= 0; i-- {
		if seen.Contains(attrs[i].Key) {
			continue
		}
		v := attrs[i].Value.Resolve()
		if v.Kind() == slog.KindGroup {
			result = append(result, slog.Group(attrs[i].Key, generic.AnySlice(UniqAttrs(v.Group()))...))
		} else {
			result = append(result, slog.Attr{Key: attrs[i].Key, Value: v})
		}
		seen.Add(attrs[i].Key)
	}
	return result
}
