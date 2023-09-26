package formatter

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.innotegrity.dev/slogx"
	"golang.org/x/exp/slog"
)

const (
	// JSONFormatterLevelAttr is the default JSON key to use when outputting the level.
	JSONFormatterLevelAttr = "@level"

	// JSONFormatterMessageAttr is the default JSON key to use when outputting the message.
	JSONFormatterMessageAttr = "@msg"

	// JSONFormatterNestedAttributeAttr is the default JSON key to use when outputting attributes, if attribute nesting
	// is enabled.
	JSONFormatterNestedAttributeAttr = "@attributes"

	// JSONFormatterSourceAttr is the default JSON key to use when outputting the source code location.
	JSONFormatterSourceAttr = "@source"

	// JSONFormatterTimeAttr is the default JSON key to use when outputting the time of the record.
	JSONFormatterTimeAttr = "@time"
)

// jsonFormatterOptionsContext can be used to retrieve the options used by the formatter from the context.
type jsonFormatterOptionsContext struct{}

// JSONFormatterOptions holds the options for the JSON formatter.
type JSONFormatterOptions struct {
	// AttrFormatter is the middleware formatting function to call to format any attribute.
	//
	// Attribute values should be resolved by the handler before formatting. Any value returned by the formatter should
	// be resolved prior to return.
	//
	// If nil, attributes remain unchanged.
	AttrFormatter FormatAttrFn

	// IgnoreAttrs is a list of regular expressions to use for matching attributes which should not be printed.
	//
	// Note that this only applies to attributes and not defined parts of the record such as time, level and the
	// message itself as they are always printed. Likewise, if NestedAttributes is true, the NestedAttributeAttr
	// is always printed and if IncludeSource is true, the source is always printed.
	//
	// If any regular expression does not compile, it is simply ignored.
	IgnoreAttrs []string

	// IncludeSource determines whether or not to include the source code location of the record in the output.
	IncludeSource bool

	// LevelAttr is the name of the JSON attribute to use for the level.
	//
	// If empty, defaults to JSONFormatterLevelAttr.
	LevelAttr string

	// LevelFormatter is the middleware formatting function to call to format the level.
	//
	// If nil, the level is printed using FormatLevelValueDefault().
	LevelFormatter FormatLevelValueFn

	// MessageAttr is the name of the JSON attribute to use for the message.
	//
	// If empty, defaults to JSONFormatterMessageAttr.
	MessageAttr string

	// MessageFormatter is the middlware formatting function to call to format the message.
	//
	// If nil, the message is printed as-is.
	MessageFormatter FormatMessageValueFn

	// NestAttributes indicates whether or not to nest the record's attributes under their own key in the JSON output.
	//
	// If true, the NestedAttributeAttr will always be part of the output even if it is empty.
	NestAttributes bool

	// NetstedAttributeAttr is the name of the JSON attribute to use if nesting attributes.
	//
	// If empty, defaults to JSONFormatterNestedAttributeAttr.
	NestedAttributeAttr string

	// SortAttrs indicates whether or not to sort attributes in the output.
	//
	// Note that this *only* affects attributes and not the time, message, source or level.
	SortAttrs bool

	// SourceAttr is the name of the JSON attribute to use for the source code location.
	//
	// If empty, defaults to JSONFormatterSourceAttr.
	SourceAttr string

	// SourceFormatter is the middleware formatting function to call to format the source code location where the record
	// was created.
	//
	// If nil, the source code location is printed using FormatSourceValueDefault().
	SourceFormatter FormatSourceValueFn

	// SpecificAttrFormatter is the middleware formatting function to call to format a specific attribute.
	//
	// The key for the map corresponds to the name of the specific attribute to format. If an attribute is nested within
	// a group, use a single period (.) to designate the group and attribute (eg: GROUP.ATTRIBUTE). Nested groups use
	// the same format (eg: GROUP1.GROUP2.ATTRIBUTE).
	//
	// Attribute values should be resolved by the handler before formatting. Any value returned by the formatter should
	// be resolved prior to return.
	//
	// If nil or if the attribute does not exist in the map, the default is to fall back to the AttrFormatter function.
	SpecificAttrFormatter map[string]FormatAttrFn

	// TimeAttr is the name of the JSON attribute to use for the time of the record.
	//
	// If empty, defaults to JSONFormatterTimeAttr.
	TimeAttr string

	// TimeFormatter is the middleware formatting function to call to the time of the record.
	//
	// If nil, the time is printed using FormatTimeValueDefault().
	TimeFormatter FormatTimeValueFn
}

// DefaultJSONFormatterOptions returns a default set of options for the JSON formatter.
func DefaultJSONFormatterOptions() JSONFormatterOptions {
	return JSONFormatterOptions{
		IgnoreAttrs: []string{},
		LevelAttr:   JSONFormatterLevelAttr,
		LevelFormatter: func(ctx context.Context, level slog.Leveler) (string, error) {
			return strings.ToLower(slogx.Level(level.Level()).String()), nil
		},
		MessageAttr:           JSONFormatterMessageAttr,
		NestAttributes:        true,
		NestedAttributeAttr:   JSONFormatterNestedAttributeAttr,
		SortAttrs:             true,
		SourceAttr:            JSONFormatterSourceAttr,
		SourceFormatter:       FormatSourceValueDefault,
		SpecificAttrFormatter: map[string]FormatAttrFn{},
		TimeAttr:              JSONFormatterTimeAttr,
		TimeFormatter:         FormatTimeValueDefault,
	}
}

// GetJSONFormatterOptionsFromContext retrieves the options from the context.
//
// If the options are not set in the context, a set of default options is returned instead.
func GetJSONFormatterOptionsFromContext(ctx context.Context) *JSONFormatterOptions {
	o := ctx.Value(jsonFormatterOptionsContext{})
	if o != nil {
		if opts, ok := o.(*JSONFormatterOptions); ok {
			return opts
		}
	}
	opts := DefaultJSONFormatterOptions()
	return &opts
}

// AddToContext adds the options to the given context and returns the new context.
func (o *JSONFormatterOptions) AddToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, jsonFormatterOptionsContext{}, o)
}

// jsonFormatter formats records for output as JSON.
type jsonFormatter struct {
	// unexported variables
	ignoredAttrPatterns []*regexp.Regexp
	options             JSONFormatterOptions
}

// DefaultJSONFormatter returns a JSON formatter with typical defaults already set.
func DefaultJSONFormatter() *jsonFormatter {
	return NewJSONFormatter(DefaultJSONFormatterOptions())
}

// NewJSONFormatter creates and returns a new JSON formatter.
func NewJSONFormatter(opts JSONFormatterOptions) *jsonFormatter {
	// set default options
	if opts.TimeAttr == "" {
		opts.TimeAttr = JSONFormatterTimeAttr
	}
	if opts.LevelAttr == "" {
		opts.LevelAttr = JSONFormatterLevelAttr
	}
	if opts.IncludeSource && opts.SourceAttr == "" {
		opts.SourceAttr = JSONFormatterSourceAttr
	}
	if opts.MessageAttr == "" {
		opts.MessageAttr = JSONFormatterMessageAttr
	}
	if opts.NestAttributes && opts.NestedAttributeAttr == "" {
		opts.NestedAttributeAttr = JSONFormatterNestedAttributeAttr
	}

	// create the formatter object
	f := &jsonFormatter{
		ignoredAttrPatterns: []*regexp.Regexp{},
		options:             opts,
	}
	for _, p := range opts.IgnoreAttrs {
		regex, err := regexp.Compile(p)
		if err == nil {
			f.ignoredAttrPatterns = append(f.ignoredAttrPatterns, regex)
		}
	}
	return f
}

// FormatRecord handles formatting the given record and outputting it into the returned buffer for consumption by a
// handler.
//
// By default, duration values in attributes are formatted using the String() function and time values are formatted
// in UTC time using the RFC3339 layout.
func (f *jsonFormatter) FormatRecord(ctx context.Context, timestamp time.Time, level slogx.Level, pc uintptr,
	msg string, attrs []slog.Attr) (*slogx.Buffer, error) {

	var err error
	var strVal string
	buf := slogx.NewBuffer()
	handlerCtx := f.options.AddToContext(ctx)

	// open the JSON
	buf.WriteByte('{')

	// write the time
	if f.options.TimeFormatter != nil {
		strVal, err = f.options.TimeFormatter(handlerCtx, level, timestamp)
	} else {
		strVal, err = FormatTimeValueDefault(handlerCtx, level, timestamp)
	}
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(buf, `"%s":"%s"`, f.options.TimeAttr, strVal)

	// write the level
	if f.options.LevelFormatter != nil {
		strVal, err = f.options.LevelFormatter(handlerCtx, level)
	} else {
		strVal, err = FormatLevelValueDefault(handlerCtx, level)
	}
	if err != nil {
		return nil, err
	}
	if buf.Len() > 2 {
		buf.WriteByte(',')
	}
	fmt.Fprintf(buf, `"%s":"%s"`, f.options.LevelAttr, strVal)

	// add source to attribute list, if enabled
	if f.options.IncludeSource {
		if f.options.SourceFormatter != nil {
			strVal, err = f.options.SourceFormatter(handlerCtx, level, pc)
		} else {
			strVal, err = FormatSourceValueDefault(handlerCtx, level, pc)
		}
		if err != nil {
			return nil, err
		}
		if buf.Len() > 2 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(buf, `"%s":"%s"`, f.options.SourceAttr, strVal)
	}

	// add message to attribute list
	if f.options.MessageFormatter != nil {
		strVal, err = f.options.MessageFormatter(handlerCtx, level, msg)
	} else {
		strVal = msg
		err = nil
	}
	if err != nil {
		return nil, err
	}
	if buf.Len() > 2 {
		buf.WriteByte(',')
	}
	fmt.Fprintf(buf, `"%s":"%s"`, f.options.MessageAttr, strVal)

	// sort attributes, if requested
	if f.options.SortAttrs {
		attrs = slogx.SortAttrs(attrs)
	}

	// loop through and print the attributes
	if f.options.NestAttributes {
		if buf.Len() > 2 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(buf, `"%s":{`, f.options.NestedAttributeAttr)
		count := 0
		for _, attr := range attrs {
			if err := f.formatAttr(handlerCtx, buf, level, "", attr.Key, attr.Value, count > 0); err != nil {
				return nil, err
			}
			count++
		}
		buf.WriteByte('}')
	} else {
		for _, attr := range attrs {
			if err := f.formatAttr(handlerCtx, buf, level, "", attr.Key, attr.Value, buf.Len() > 2); err != nil {
				return nil, err
			}
		}
	}

	// close the JSON
	buf.WriteString("}\n")
	return buf, nil
}

// formatAttr formats the given attribute key and value and returns the resulting string to print to the buffer.
//
// By default, duration values in attributes are formatted using the String() function and time values are formatted
// in UTC time using the RFC3339 layout.
func (f jsonFormatter) formatAttr(ctx context.Context, buf *slogx.Buffer, level slog.Leveler, group, attrKey string,
	attrValue slog.Value, writeComma bool) error {

	// create the full key path with the group
	groupWithKey := attrKey
	if group != "" {
		groupWithKey = fmt.Sprintf("%s.%s", group, attrKey)
	}

	// ignore the given attribute if the group/key matches
	for _, p := range f.ignoredAttrPatterns {
		if p.MatchString(groupWithKey) {
			return nil
		}
	}

	// format the attribute using any formatter functions first
	formattedKey := attrKey
	formattedValue := attrValue
	var err error
	if fn, ok := f.options.SpecificAttrFormatter[groupWithKey]; ok && fn != nil {
		formattedKey, formattedValue, err = fn(ctx, level, group, formattedKey, formattedValue)
		if err != nil {
			return err
		}
	} else if f.options.AttrFormatter != nil {
		formattedKey, formattedValue, err = f.options.AttrFormatter(ctx, level, group, formattedKey, formattedValue)
		if err != nil {
			return err
		}
	}

	// format the key/value
	if writeComma {
		buf.WriteByte(',')
	}
	switch formattedValue.Kind() {
	case slog.KindBool:
		fmt.Fprintf(buf, `"%s":%t`, formattedKey, formattedValue.Bool())
	case slog.KindString:
		fmt.Fprintf(buf, `"%s":"%s"`, formattedKey, formattedValue.String())
	case slog.KindDuration:
		fmt.Fprintf(buf, `"%s":"%s"`, formattedKey, formattedValue.Duration().String())
	case slog.KindTime:
		fmt.Fprintf(buf, `"%s":"%s"`, formattedKey, formattedValue.Time().UTC().Format(time.RFC3339))
	case slog.KindFloat64:
		fmt.Fprintf(buf, `"%s":%f`, formattedKey, formattedValue.Float64())
	case slog.KindInt64:
		fmt.Fprintf(buf, `"%s":%d`, formattedKey, formattedValue.Int64())
	case slog.KindUint64:
		fmt.Fprintf(buf, `"%s":%d`, formattedKey, formattedValue.Uint64())
	case slog.KindGroup:
		fmt.Fprintf(buf, `"%s":{`, formattedKey)
		count := 0
		for _, attr := range formattedValue.Group() {
			if err := f.formatAttr(ctx, buf, level, groupWithKey, attr.Key, attr.Value, count > 0); err != nil {
				return err
			}
			count++
		}
		buf.WriteByte('}')
	default:
		marshalled, err := json.Marshal(formattedValue.Any())
		if err != nil {
			return err
		}
		fmt.Fprintf(buf, `"%s":%s`, formattedKey, marshalled)
	}
	return nil
}
