package slogx

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"go.innotegrity.dev/errorx"
	"go.innotegrity.dev/generic"
	"golang.org/x/exp/slices"
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

// Err returns an Attr for an error value.
func Err(key string, value error) slog.Attr {
	if value == nil {
		return slog.Attr{
			Key:   key,
			Value: slog.AnyValue(nil),
		}
	}
	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(value.Error()),
	}
}

// ErrX returns an Attr for an extended error value.
func ErrX(key string, value errorx.Error) slog.Attr {
	if value == nil {
		return slog.Attr{
			Key:   key,
			Value: slog.AnyValue(nil),
		}
	}

	// add the core attributes
	attrs := []any{
		slog.Int("code", value.Code()),
		slog.String("error", value.Error()),
	}
	err := value.InternalError()
	if err != nil {
		attrs = append(attrs, slog.String("internal_error", err.Error()))
	}
	// add any attributes from the error
	errorAttrs := []any{}
	for k, v := range value.Attrs() {
		errorAttrs = append(errorAttrs, slog.Any(k, v))
	}
	if len(errorAttrs) > 0 {
		attrs = append(attrs, slog.Group("attributes", errorAttrs...))
	}

	// add nested errors
	nestedErrs := []any{}
	for i, ne := range value.NestedErrors() {
		nestedErrs = append(nestedErrs, ErrX(fmt.Sprintf("%03d", i+1), ne))
	}
	if len(nestedErrs) > 0 {
		attrs = append(attrs, slog.Group("nested_errors", nestedErrs...))
	}
	v := slog.Group(key, attrs...)
	return v
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

// HttpRequest returns an Attr for an HTTP request object.
func HttpRequest(key string, req *http.Request, sensitiveHeaders []string, sensitiveQueryParams []string) slog.Attr {
	if req == nil {
		return slog.Attr{
			Key:   key,
			Value: slog.AnyValue(nil),
		}
	}

	// add headers
	headerAttrs := []any{}
	for header, value := range req.Header {
		v := strings.Join(value, ",")
		if slices.Contains(sensitiveHeaders, header) {
			v = "************"
		}
		headerAttrs = append(headerAttrs, slog.String(header, v))
	}

	// add query parameters
	queryAttrs := []any{}
	for key, value := range req.URL.Query() {
		v := strings.Join(value, ",")
		if slices.Contains(sensitiveQueryParams, key) {
			v = "************"
		}
		queryAttrs = append(queryAttrs, slog.String(key, v))
	}

	if req.URL.Path[0] == '/' {
		req.URL.Path = req.URL.Path[1:]
	}
	return slog.Group(
		key,
		slog.String("host", req.Host),
		slog.String("method", req.Method),
		slog.String("user_agent", req.UserAgent()),
		slog.String("url", fmt.Sprintf("%s://%s/%s", req.URL.Scheme, req.URL.Host, req.URL.Path)),
		slog.Group("url",
			slog.String("scheme", req.URL.Scheme),
			slog.String("host", req.URL.Host),
			slog.String("path", req.URL.Path),
			slog.String("fragment", req.URL.Fragment),
			slog.Group("query", queryAttrs...),
		),
		slog.Group("headers", headerAttrs...),
	)
}

// HttpResponse returns an Attr for an HTTP response object.
func HttpResponse(key string, resp *http.Response, sensitiveHeaders []string) slog.Attr {
	if resp == nil {
		return slog.Attr{
			Key:   key,
			Value: slog.AnyValue(nil),
		}
	}

	// add headers
	headerAttrs := []any{}
	for header, value := range resp.Header {
		v := strings.Join(value, ",")
		if slices.Contains(sensitiveHeaders, header) {
			v = "************"
		}
		headerAttrs = append(headerAttrs, slog.String(header, v))
	}

	return slog.Group(
		key,
		slog.String("status", resp.Status),
		slog.Int("status_code", resp.StatusCode),
		slog.String("proto", resp.Proto),
		slog.Int64("content_length", resp.ContentLength),
		slog.Bool("uncompressed", resp.Uncompressed),
		slog.Group("headers", headerAttrs...),
	)
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
