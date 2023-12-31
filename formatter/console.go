package formatter

import (
	"context"
	"encoding"
	"fmt"
	"regexp"
	"strings"
	"time"

	"log/slog"

	"github.com/fatih/color"
	"go.innotegrity.dev/generic"
	"go.innotegrity.dev/runtimex"
	"go.innotegrity.dev/slogx"
)

const (
	// ConsoleFormatterAttrsPart is the part used by PartOrder for all attributes.
	ConsoleFormatterAttrsPart = "attrs"

	// ConsoleFormatterLevelPart is the part used for PartOrder for the message level.
	ConsoleFormatterLevelPart = "level"

	// ConsoleFormatterMessagePart is the part used for PartOrder for the message itself.
	ConsoleFormatterMessagePart = "message"

	// ConsoleFormatterSourcePart is the part used for PartOrder for the source code location for the record.
	ConsoleFormatterSourcePart = "source"

	// ConsoleFormatterTimePart is the part used for PartOrder for the time of the message/record.
	ConsoleFormatterTimePart = "time"

	// consoleFormatterAttrPart is the prefix prepended to a specific attribute in PartOrder.
	consoleFormatterAttrPart = "attr:"

	// consoleFormatterAttrRegexPart is the prefix prepended to an attribute regular expression in PartOrder.
	consoleFormatterAttrRegexPart = "attrexpr:"
)

// ConsoleFormatterPart is just a string.
type ConsoleFormatterPart string

// IsRegexAttrPart indicates whether or not this part is for attributes matching a regular expression.
func (p ConsoleFormatterPart) IsRegexAttr() bool {
	return strings.HasPrefix(string(p), consoleFormatterAttrRegexPart)
}

// IsSpecificAttr indicates whether or not this part is for a specific attribute.
func (p ConsoleFormatterPart) IsSpecificAttr() bool {
	return strings.HasPrefix(string(p), consoleFormatterAttrPart)
}

// GetAttrRegex returns the regular expression to use when matching attributes attribute if this part contains a
// regular expression. Otherwise it returns an empty string.
func (p ConsoleFormatterPart) GetAttrRegex() string {
	if !p.IsRegexAttr() {
		return ""
	}
	return string(p)[len(consoleFormatterAttrRegexPart):]
}

// GetAttr returns the name of the specific attribute if this part is for a specific attribute. Otherwise it returns
// an empty string.
func (p ConsoleFormatterPart) GetAttr() string {
	if !p.IsSpecificAttr() {
		return ""
	}
	return string(p)[len(consoleFormatterAttrPart):]
}

// ConsoleFormatterAttrPart is the part used by PartOrder for a specific attribute.
func ConsoleFormatterAttrPart(attrKey string) ConsoleFormatterPart {
	return ConsoleFormatterPart(fmt.Sprintf("%s%s", consoleFormatterAttrPart, attrKey))
}

// ConsoleFormatterAttrRegexPart is the part used by PartOrder for an attribute regular expression.
func ConsoleFormatterAttrRegexPart(attrKey string) ConsoleFormatterPart {
	return ConsoleFormatterPart(fmt.Sprintf("%s%s", consoleFormatterAttrRegexPart, attrKey))
}

// consoleFormatterOptionsContext can be used to retrieve the options used by the formatter from the context.
type consoleFormatterOptionsContext struct{}

// ConsoleFormatterOptions holds the options for the console formatter.
type ConsoleFormatterOptions struct {
	// AttrFormatter is the middleware formatting function to call to format any attribute.
	//
	// Attribute values should be resolved by the handler before formatting. Any value returned by the formatter should
	// be resolved prior to return.
	//
	// If nil, attributes are simply printed unchanged as key=value.
	AttrFormatter FormatAttrFn

	// EnableColor determines whether or not to enable colorized output.
	EnableColor bool

	// IgnoreAttrs is a list of regular expressions to use for matching attributes which should not be printed.
	//
	// Note that this only applies to attributes and not defined parts like the level, message, source or time. If you
	// want to ignore those, simply leave them out of the PartOrder array.
	//
	// If any regular expression does not compile, it is simply ignored.
	IgnoreAttrs []string

	// LevelFormatter is the middleware formatting function to call to format the level.
	//
	// If nil, the level is printed using FormatLevelValueDefault().
	LevelFormatter FormatLevelValueFn

	// MessageFormatter is the middlware formatting function to call to format the message.
	//
	// If nil, the message is printed as-is.
	MessageFormatter FormatMessageValueFn

	// PartOrder is the order in which to print the various parts of the message.
	//
	// The following values are valid for the string:
	// ConsoleFormatterAttrPart - specific attribute to print; use a single period (.) to separate group name from
	//                            attribute name if the attribute is nested within a group
	// ConsoleFormatterAttrRegexPart - attributes matching the regular expression to print; use an escaped period (\.)
	//                                 to separate group name from attribute name if the attribute is nested within a
	//                                 group; if the regex does not compile, it is ignored
	// ConsoleFormatterAttrsPart - all attributes
	// ConsoleFormatterLevelPart - the log level of the message from the record as a string
	// ConsoleFormatterMessagePart - the message from the record
	// ConsoleFormatterSourcePart - the location at which the message was logged
	// ConsoleFormatterTimePart - the time of the message from the record
	//
	// If any other string is specified other than those above, it is simply printed as-is.
	//
	// If an attribute is specified but not present in the record, it is ignored.
	//
	// By default the format will be "TimePart LevelPart SourcePart > MessagePart AttrsPart".
	PartOrder []ConsoleFormatterPart

	// PartSeparator is how to separate the parts when they're being printed. Must be at least 1 character.
	//
	// By default, a space is used.
	PartSeparator string

	// SortAttributes determines whether or not to sort attributes in the output.
	//
	// Note that this *only* affects the output for ConsoleFormatterAttrsPart.
	SortAttributes bool

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

	// TimeFormatter is the middleware formatting function to call to the time of the record.
	//
	// If nil, the time is printed using FormatTimeValueDefault().
	TimeFormatter FormatTimeValueFn

	// UniqueAttributesOnly indicates whether or not to only print unique attributes.
	//
	// If multiple attributes are present within the same group with the same key name, only the latest attribute
	// will be printed.
	//
	// For example:
	//   logger := logger.With(slog.String("k1", "v1"))
	//   logger.Warn("this is a warning message", slog.String("k1", "v2"))
	//
	// If this value is true and the code above is used, only {"k1":"v2"} would be shown as an attribute.
	//
	// The only exception to this rule is that attributes which are not nested and overlap with the time, level, source
	// or message keys will be ignored as time, level, source and message are core parts of the output.
	UniqueAttributesOnly bool
}

// ConsoleFormatterOptionsFromContext retrieves the formatter options from the context.
//
// If the options are not set in the context, a set of default options is returned instead.
func ConsoleFormatterOptionsFromContext(ctx context.Context) *ConsoleFormatterOptions {
	o := ctx.Value(consoleFormatterOptionsContext{})
	if o != nil {
		if opts, ok := o.(*ConsoleFormatterOptions); ok {
			return opts
		}
	}
	opts := DefaultConsoleFormatterOptions()
	return &opts
}

// ContextWithConsoleFormatterOptions adds the options to the given context and returns the new context.
func ContextWithConsoleFormatterOptions(ctx context.Context, opts ConsoleFormatterOptions) context.Context {
	return context.WithValue(ctx, consoleFormatterOptionsContext{}, &opts)
}

// DefaultConsoleFormatterOptions returns a default set of options for the console formatter.
func DefaultConsoleFormatterOptions() ConsoleFormatterOptions {
	return ConsoleFormatterOptions{
		IgnoreAttrs:    []string{},
		LevelFormatter: FormatLevelValueDefault,
		PartOrder: []ConsoleFormatterPart{
			ConsoleFormatterTimePart,
			ConsoleFormatterLevelPart,
			ConsoleFormatterSourcePart,
			">",
			ConsoleFormatterMessagePart,
			ConsoleFormatterAttrPart("error"),
			ConsoleFormatterAttrRegexPart(`error\..*`),
			ConsoleFormatterAttrsPart,
		},
		PartSeparator:   " ",
		SortAttributes:  true,
		SourceFormatter: FormatSourceValueDefault,
		TimeFormatter: func(ctx context.Context, level slog.Leveler, t time.Time) (string, error) {
			return t.Local().Format("03:04:05PM"), nil
		},
		UniqueAttributesOnly: true,
	}
}

// consoleFormatter formats records for output to a console such as stdout, stderr or even a file.
type consoleFormatter struct {
	// unexported variables
	ignoredAttrPatterns []*regexp.Regexp
	options             ConsoleFormatterOptions
	willPrintAttrs      bool
}

// DefaultConsoleFormatter returns a console formatter with typical defaults already set.
func DefaultConsoleFormatter(colorize bool) *consoleFormatter {
	options := DefaultConsoleFormatterOptions()
	if colorize {
		options.EnableColor = true
		options.AttrFormatter = ColorizeAttrFormatter
		options.LevelFormatter = ColorizeLevelFormatter
		options.PartOrder = []ConsoleFormatterPart{
			ConsoleFormatterTimePart,
			ConsoleFormatterLevelPart,
			ConsoleFormatterSourcePart,
			ConsoleFormatterPart(color.New(color.FgHiWhite).Sprint(">")),
			ConsoleFormatterMessagePart,
			ConsoleFormatterAttrPart("error"),
			ConsoleFormatterAttrRegexPart(`error\..*`),
			ConsoleFormatterAttrsPart,
		}
		options.SourceFormatter = ColorizeSourceFormatter
		options.SpecificAttrFormatter = map[string]FormatAttrFn{
			"error": ColorizeErrorAttrFormatter,
		}
	}
	return NewConsoleFormatter(options)
}

// NewConsoleFormatter creates and returns a new console formatter.
func NewConsoleFormatter(opts ConsoleFormatterOptions) *consoleFormatter {
	// set default options
	if len(opts.PartOrder) == 0 {
		opts.PartOrder = []ConsoleFormatterPart{
			ConsoleFormatterTimePart,
			ConsoleFormatterLevelPart,
			ConsoleFormatterSourcePart,
			">",
			ConsoleFormatterMessagePart,
			ConsoleFormatterAttrPart("error"),
			ConsoleFormatterAttrRegexPart(`error\..*`),
			ConsoleFormatterAttrsPart,
		}
	}
	if opts.PartSeparator == "" {
		opts.PartSeparator = " "
	}

	// create the formatter object
	f := &consoleFormatter{
		ignoredAttrPatterns: []*regexp.Regexp{},
		options:             opts,
		willPrintAttrs:      false,
	}
	for _, p := range opts.PartOrder {
		if p.IsSpecificAttr() || p.IsRegexAttr() || p == ConsoleFormatterAttrsPart {
			f.willPrintAttrs = true
		}
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
func (f *consoleFormatter) FormatRecord(ctx context.Context, timestamp time.Time, level slogx.Level, pc uintptr,
	msg string, attrs []slog.Attr) (*slogx.Buffer, error) {

	var err error
	var strVal string
	buf := slogx.NewBuffer()
	formatterCtx := ContextWithConsoleFormatterOptions(ctx, f.options)

	// flatten attributes
	var attrMap map[string]slog.Value
	if f.willPrintAttrs {
		if f.options.SortAttributes {
			attrs = slogx.SortAttrs(attrs)
		}
		attrs = slogx.FlattenAttrs(attrs)
	}

	// now let's actually print the parts out
	lastBufLen := 0
	printedAttrs := generic.NewSet[string]()
	for _, part := range f.options.PartOrder {
		// only print the parts separator if we actually printed something before
		currentBufLen := buf.Len()
		if currentBufLen > 0 && currentBufLen != lastBufLen {
			fmt.Fprintf(buf, "%s", f.options.PartSeparator)
			currentBufLen = buf.Len()
		}
		lastBufLen = currentBufLen

		switch part {
		case ConsoleFormatterAttrsPart:
			if err = f.printAttrs(formatterCtx, buf, level, attrs, printedAttrs); err != nil {
				return nil, err
			}

		case ConsoleFormatterLevelPart:
			if f.options.LevelFormatter != nil {
				strVal, err = f.options.LevelFormatter(formatterCtx, level)
			} else {
				strVal, err = FormatLevelValueDefault(formatterCtx, level)
			}
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(buf, "%s", strVal)

		case ConsoleFormatterMessagePart:
			if f.options.MessageFormatter != nil {
				strVal, err = f.options.MessageFormatter(formatterCtx, level, msg)
			} else {
				strVal = msg
				err = nil
			}
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(buf, "%s", strVal)

		case ConsoleFormatterSourcePart:
			if f.options.SourceFormatter != nil {
				strVal, err = f.options.SourceFormatter(formatterCtx, level, pc)
			} else {
				strVal, err = FormatSourceValueDefault(formatterCtx, level, pc)
			}
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(buf, "%s", strVal)

		case ConsoleFormatterTimePart:
			if f.options.TimeFormatter != nil {
				strVal, err = f.options.TimeFormatter(formatterCtx, level, timestamp)
			} else {
				strVal, err = FormatTimeValueDefault(formatterCtx, level, timestamp)
			}
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(buf, "%s", strVal)

		default:
			if len(attrMap) == 0 {
				attrMap = slogx.ToAttrMap(attrs)
			}
			if attr := part.GetAttr(); attr != "" { // specific attribute
				if val, ok := attrMap[attr]; ok {
					if err = f.printAttr(formatterCtx, buf, level, attr, val, printedAttrs); err != nil {
						return nil, err
					}
				}
			} else if attrRegex := part.GetAttrRegex(); attrRegex != "" { // attribute regex
				regex, err := regexp.Compile(attrRegex)
				if err == nil {
					for attr, val := range attrMap {
						currentBufLen = buf.Len()
						if currentBufLen > 0 && currentBufLen != lastBufLen {
							fmt.Fprintf(buf, "%s", f.options.PartSeparator)
							currentBufLen = buf.Len()
						}
						lastBufLen = currentBufLen
						if regex.MatchString(attr) {
							if err := f.printAttr(formatterCtx, buf, level, attr, val, printedAttrs); err != nil {
								return nil, err
							}
						}
					}
				}
			} else { // raw string without formatting
				fmt.Fprint(buf, part)
			}
		}
	}

	// finally - write the message
	buf.WriteByte('\n')
	return buf, nil
}

// IsColorized returns whether or not the formatter is enabled for colorizing the output.
func (f consoleFormatter) IsColorized() bool {
	return f.options.EnableColor
}

// printAttr prints the given
func (f consoleFormatter) printAttr(ctx context.Context, buf *slogx.Buffer, level slog.Leveler, attrKey string,
	attrValue slog.Value, printedAttrs generic.Set[string]) error {

	// already printed the given key
	if printedAttrs.Contains(attrKey) {
		return nil
	}

	// ignore the attribute if the key matches
	for _, p := range f.ignoredAttrPatterns {
		if p.MatchString(attrKey) {
			return nil
		}
	}

	// extract the group name and attribute from the key
	group := ""
	actualAttrKey := attrKey
	groupIndex := strings.LastIndex(attrKey, ".")
	if groupIndex != -1 {
		group = attrKey[:groupIndex]
		actualAttrKey = attrKey[groupIndex+1:]
	}

	// format the attribute using any formatter functions first
	formattedKey := attrKey
	formattedValue := attrValue.Resolve()
	var err error
	if fn, ok := f.options.SpecificAttrFormatter[attrKey]; ok && fn != nil {
		formattedKey, formattedValue, err = fn(ctx, level, group, actualAttrKey, formattedValue)
		if err != nil {
			return err
		}
	} else if f.options.AttrFormatter != nil {
		formattedKey, formattedValue, err = f.options.AttrFormatter(ctx, level, group, actualAttrKey, formattedValue)
		if err != nil {
			return err
		}
	}

	// format the key/value
	switch formattedValue.Kind() {
	case slog.KindBool:
		fmt.Fprintf(buf, "%s=%t", formattedKey, formattedValue.Bool())
	case slog.KindString:
		fmt.Fprintf(buf, "%s=%s", formattedKey, formattedValue.String())
	case slog.KindDuration:
		fmt.Fprintf(buf, "%s=%s", formattedKey, formattedValue.Duration().String())
	case slog.KindTime:
		fmt.Fprintf(buf, "%s=%s", formattedKey, formattedValue.Time().UTC().Format(time.RFC3339))
	case slog.KindFloat64:
		fmt.Fprintf(buf, "%s=%f", formattedKey, formattedValue.Float64())
	case slog.KindInt64:
		fmt.Fprintf(buf, "%s=%d", formattedKey, formattedValue.Int64())
	case slog.KindUint64:
		fmt.Fprintf(buf, "%s=%d", formattedKey, formattedValue.Uint64())
	case slog.KindGroup:
		for i, attr := range formattedValue.Group() {
			if i > 0 {
				buf.WriteString(f.options.PartSeparator)
			}
			groupKey := fmt.Sprintf("%s.%s", attrKey, attr.Key)
			if err := f.printAttr(ctx, buf, level, groupKey, attr.Value, printedAttrs); err != nil {
				return err
			}
			printedAttrs.Add(groupKey)
		}
	default:
		if tm, ok := formattedValue.Any().(encoding.TextMarshaler); ok {
			output, err := tm.MarshalText()
			if err != nil {
				return err
			}
			fmt.Fprintf(buf, "%s=%s", formattedKey, string(output))
		} else {
			fmt.Fprintf(buf, "%s=%+v", formattedKey, formattedValue.Any())
		}
	}
	printedAttrs.Add(attrKey)
	return nil
}

// printAttrs prints the given list of attributes to the buffer.
func (f consoleFormatter) printAttrs(ctx context.Context, buf *slogx.Buffer, level slog.Leveler, attrs []slog.Attr,
	printedAttrs generic.Set[string]) error {

	lastBufLen := buf.Len()
	for _, attr := range attrs {
		// already printed the given key
		if printedAttrs.Contains(attr.Key) {
			continue
		}

		// only print the parts separator if we actually printed something before
		currentBufLen := buf.Len()
		if currentBufLen > 0 && currentBufLen != lastBufLen {
			buf.WriteString(f.options.PartSeparator)
			currentBufLen = buf.Len()
		}
		lastBufLen = currentBufLen

		// print the attribute
		if err := f.printAttr(ctx, buf, level, attr.Key, attr.Value, printedAttrs); err != nil {
			return err
		}
	}
	return nil
}

// ColorizeAttrFormatter is a customized formatter for colorizing attribute keys.
func ColorizeAttrFormatter(ctx context.Context, level slog.Leveler, group, attrKey string,
	attrValue slog.Value) (string, slog.Value, error) {

	c := color.New(color.FgHiBlue)
	if group != "" {
		attrKey = fmt.Sprintf("%s.%s", group, attrKey)
	}
	return c.Sprint(attrKey), attrValue, nil
}

// ColorizeErrorAttrFormatter is a customized formatter for colorizing error keys.
func ColorizeErrorAttrFormatter(ctx context.Context, level slog.Leveler, group, attrKey string,
	attrValue slog.Value) (string, slog.Value, error) {

	c := color.New(color.FgHiRed)
	if group != "" {
		attrKey = fmt.Sprintf("%s.%s", group, attrKey)
	}
	return c.Sprint(attrKey), attrValue, nil
}

// ColorizeLevelFormatter is a customized formatter for colorizing levels.
func ColorizeLevelFormatter(ctx context.Context, level slog.Leveler) (string, error) {
	var levelStr string
	var c *color.Color
	switch level {
	case slogx.LevelTrace:
		c = color.New(color.FgBlue)
		levelStr = slogx.Level(level.Level()).ShortString()
	case slogx.LevelDebug:
		c = color.New(color.FgHiMagenta)
		levelStr = slogx.Level(level.Level()).ShortString()
	case slogx.LevelInfo:
		c = color.New(color.FgHiGreen)
		levelStr = slogx.Level(level.Level()).ShortString()
	case slogx.LevelNotice:
		c = color.New(color.FgYellow)
		levelStr = slogx.Level(level.Level()).ShortString()
	case slogx.LevelWarn:
		c = color.New(color.FgHiYellow)
		levelStr = slogx.Level(level.Level()).ShortString()
	case slogx.LevelError:
		c = color.New(color.FgHiRed)
		levelStr = slogx.Level(level.Level()).ShortString()
	case slogx.LevelFatal:
		c = color.New(color.FgHiRed, color.BgWhite)
		levelStr = slogx.Level(level.Level()).ShortString()
	case slogx.LevelPanic:
		c = color.New(color.FgHiRed, color.BgHiYellow)
		levelStr = slogx.Level(level.Level()).ShortString()
	default:
		c = color.New(color.FgHiCyan)
		levelStr = strings.ToUpper(level.Level().String())
		if len(levelStr) > 3 {
			levelStr = levelStr[0:3]
		}
	}
	return c.Sprint(levelStr), nil
}

// ColorizeSourceFormatter is a customized formatter for colorizing the source file and line.
func ColorizeSourceFormatter(ctx context.Context, level slog.Leveler, pc uintptr) (string, error) {
	f := runtimex.FrameFromPC(pc)
	return color.New(color.FgHiWhite).Sprintf("%s", f), nil
}
