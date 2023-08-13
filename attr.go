package slogx

import (
	"fmt"
	"net/http"
	"strings"

	"go.innotegrity.dev/errorx"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

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
		nestedErrs = append(nestedErrs, ErrX(fmt.Sprintf("%d", i+1), ne))
	}
	if len(nestedErrs) > 0 {
		attrs = append(attrs, slog.Group("nested_errors", nestedErrs...))
	}
	v := slog.Group(key, attrs...)
	return v
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
