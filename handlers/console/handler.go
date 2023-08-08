package console

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"go.innotegrity.dev/slogx"
	"golang.org/x/exp/slog"
)

const (
	// AttrKey is the key used for PartsOrder and PartsFormatter for a specific attribute. You must supply the
	// attribute name via fmt.Sprintf(AttrKey, ATTRIBUTE_NAME) where ATTRIBUTE_NAME is the actual attribute.
	AttrKey = "attr:%s"

	// AttrsKey is the key used for PartsOrder and PartsFormatter for all attributes in no particular order.
	AttrsKey = "attrs"

	// AttrsAlphaKey is the key used for PartsOrder and PartsFormatter for all attributes in alphabetical order.
	AttrsAlphaKey = "attrs_alpha"

	// CallerKey is the key used for PartsOrder and PartsFormatter for the filename and line.
	CallerKey = "caller"

	// DividerKey is the key used for PartsOrder and PartsFormatter for the divider string supplied in the options.
	DividerKey = "divider"

	// LevelKey is the key used for PartsOrder and PartsFormatter for the message level.
	LevelKey = "level"

	// MessageKey is the key used for PartsOrder and PartsFormatter for the message itself.
	MessageKey = "message"

	// TimeKey is the key used for PartsOrder and PartsFormatter for the time of the message/record.
	TimeKey = "time"
)

// FormatterFn is the function called to format a specific attribute in the record.
type FormatterFn func(context.Context, string, slog.Record) string

// LevelColor stores the foreground and background for colorizing a level.
type LevelColor struct {
	// FG is the foreground color for the level.
	FG int

	// BG is the background color for the level.
	BG int
}

// Options holds the options for the handler.
type Options struct {
	// Divider is the divider to string to use for the divider key.
	//
	// By default, the divider will be set to a single pipe (|) character.
	Divider string

	// Level is the minimum log level to write to the handler.
	//
	// By default, the level will be set to slog.LevelInfo if not supplied.
	Level slog.Leveler

	// NoColor disables any colorized output.
	NoColor bool

	// PartsFormatter is a map of how to format the various parts for the message.
	//
	// By default, the following formats are used by the handler:
	// AttrsKey - key=value with key colorized in bright blue
	// CallerKey - filename:line colorized in bright white
	// DividerKey - the divider string specified (default is a single pipe (|) character) colorized in bright blue
	// LevelKey - the level as 3 letter colorized string in uppercase
	//				    slogx.LevelTrace - TRC in cyan
	//   					slog.LevelDebug - DBG in magenta
	//	   			 	slog.LevelInfo - INF in bright blue
	//          	slogx.LevelNotice - NOT in bright yellow (orange)
	//				 		slog.LevelWarn - WRN in yellow
	//				 		slog.LevelError - ERR in bright red
	//				 		slogx.LevelFatal - FTL in bright red with white background
	//				 		slogx.LevelPanic - PNC in bright red with yellow background
	//				 		any other level - first 3 letters capitalized in bright white
	// MessageKey - just the message itself
	// TimeKey - HH:MM followed by AM/PM in the local time zone colorized in white
	//
	// To specify a formatter for either AttrsKey or AttrsAlphaKey, only use AttrsKey.
	//
	// To define a formatted for a particular attribute, use Sprintf(AttrKey, NAME) where NAME is the specific attribute
	// to use and use it as the key to the PartsFormatter. If a particular attribute does not have its own formatter
	// defined, it is simply printed using the attribute formatter. If an attribute is nested within a group, use a
	// single period (.) to separate the group name from the attribute name.
	//
	// See PartsOrder for valid part names.
	PartsFormatter map[string]FormatterFn

	// PartsOrder is the order in which to print the various parts of the message.
	//
	// The following values are valid for the string:
	// AttrKey - specific attribute - use Sprintf(AttrKey, NAME) where NAME is the specific attribute to use; use
	//           a single period (.) to separate group name from attribute name if the attribute is nested within a
	//           group
	// AttrsKey - all attributes in any order
	// AttrsAlphaKey - all attributes in alphabetical order
	// CallerKey - the filename and line where the message was logged
	// DividerKey - an arbitrary divider string
	// LevelKey - the log level of the message
	// MessageKey - the level itself
	// TimeKey - the time of the message
	//
	// If any other string is specified other than those above, it is simply printed as-is.
	//
	// Any attribute will only be output once.  If a specific attribute is output using the AttrKey, it will not
	// be repeated later if AttrsKey or AttrsAlphaKey is specified. Likewise, AttrsKey and AttrsAlphaKey should
	// never both be specified as only the first one will be used.
	//
	// If an attribute is specified but not present in the record, it is ignored.
	//
	// By default the format will be "TimeKey LevelKey CallerKey DividerKey MessageKey AttrsAlphaKey".
	PartsOrder []string

	// PartsSeparator is how to separate the parts when they're being printed. Must be at least 1 character.
	//
	// By default, a space is used.
	PartsSeparator string

	// Writer is where to write the output to.
	//
	// By default, messages are written to os.Stdout if not supplied.
	Writer io.Writer
}

// NewHandler creates a new handler.
func (o Options) NewHandler() slog.Handler {
	if o.Divider == "" {
		o.Divider = "|"
	}
	if o.Level == nil {
		o.Level = slog.LevelInfo
	}
	if o.PartsFormatter == nil {
		o.PartsFormatter = map[string]FormatterFn{}
	}
	if len(o.PartsOrder) == 0 {
		o.PartsOrder = []string{
			TimeKey,
			LevelKey,
			CallerKey,
			DividerKey,
			MessageKey,
			AttrsAlphaKey,
		}
	}
	if o.PartsSeparator == "" {
		o.PartsSeparator = " "
	}
	if o.Writer == nil {
		o.Writer = os.Stdout
	}
	if (o.Writer == os.Stdout || o.Writer == os.Stderr) && !o.NoColor {
		o.Writer = colorable.NewColorable(o.Writer.(*os.File))
	}
	return &Handler{
		options: o,
		attrs:   []slog.Attr{},
		groups:  []string{},
	}
}

// Handler is a log handler that writes records to an io.Writer, typically a console in a specified format.
type Handler struct {
	options Options
	attrs   []slog.Attr
	groups  []string
}

// Enabled determines whether or not the given level is enabled in this handler.
func (h Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.options.Level.Level()
}

// Handle actually handles writing the record to the output writer.
func (h Handler) Handle(ctx context.Context, r slog.Record) error {
	buf := bytes.NewBuffer(nil)
	printedAttrs := map[string]bool{}
	lastBufLen := 0

partsloop:
	for _, part := range h.options.PartsOrder {
		// only print the parts separator if we actually printed something before
		currentBufLen := buf.Len()
		if currentBufLen > 0 && currentBufLen != lastBufLen {
			fmt.Fprintf(buf, "%s", h.options.PartsSeparator)
			currentBufLen = buf.Len()
		}
		lastBufLen = currentBufLen

		switch part {
		case AttrsAlphaKey:
			keyOrder, attrs := h.resolveRecordAttrs(ctx, buf, r, true)
			h.printAttrs(ctx, r, buf, keyOrder, attrs, printedAttrs)

		case AttrsKey:
			keyOrder, attrs := h.resolveRecordAttrs(ctx, buf, r, false)
			h.printAttrs(ctx, r, buf, keyOrder, attrs, printedAttrs)

		case CallerKey:
			if fn, ok := h.options.PartsFormatter[CallerKey]; ok && fn != nil {
				fmt.Fprintf(buf, "%s", fn(ctx, CallerKey, r))
				continue partsloop
			}
			frames := runtime.CallersFrames([]uintptr{r.PC})
			frame, _ := frames.Next()
			if cwd, err := os.Getwd(); err == nil {
				if rel, err := filepath.Rel(cwd, frame.File); err == nil {
					frame.File = rel
				}
			}
			h.colorize(buf, color.New(color.FgHiWhite), fmt.Sprintf("%s:%d", frame.File, frame.Line))

		case DividerKey:
			if fn, ok := h.options.PartsFormatter[DividerKey]; ok && fn != nil {
				fmt.Fprintf(buf, "%s", fn(ctx, LevelKey, r))
				continue partsloop
			}
			h.colorize(buf, color.New(color.FgHiBlue), "|")

		case LevelKey:
			if fn, ok := h.options.PartsFormatter[LevelKey]; ok && fn != nil {
				fmt.Fprintf(buf, "%s", fn(ctx, LevelKey, r))
				continue partsloop
			}
			switch r.Level {
			case slogx.LevelTrace:
				h.colorize(buf, color.New(color.FgCyan), "TRC")
			case slog.LevelDebug:
				h.colorize(buf, color.New(color.FgMagenta), "DBG")
			case slog.LevelInfo:
				h.colorize(buf, color.New(color.FgHiBlue), "INF")
			case slogx.LevelNotice:
				h.colorize(buf, color.New(color.FgHiYellow), "NOT")
			case slog.LevelWarn:
				h.colorize(buf, color.New(color.FgYellow), "WRN")
			case slog.LevelError:
				h.colorize(buf, color.New(color.FgHiRed), "ERR")
			case slogx.LevelFatal:
				h.colorize(buf, color.New(color.FgHiRed, color.BgWhite), "FTL")
			case slogx.LevelPanic:
				h.colorize(buf, color.New(color.FgHiRed, color.BgHiYellow), "PAN")
			default:
				str := strings.ToUpper(r.Level.String())
				if len(str) > 3 {
					str = str[0:3]
				}
				h.colorize(buf, color.New(color.FgHiWhite), str)
			}

		case MessageKey:
			if fn, ok := h.options.PartsFormatter[MessageKey]; ok && fn != nil {
				fmt.Fprintf(buf, "%s", fn(ctx, MessageKey, r))
				continue partsloop
			}
			h.colorize(buf, color.New(color.FgWhite), r.Message)

		case TimeKey:
			if fn, ok := h.options.PartsFormatter[TimeKey]; ok && fn != nil {
				fmt.Fprintf(buf, "%s", fn(ctx, TimeKey, r))
				continue partsloop
			}
			h.colorize(buf, color.New(color.FgWhite), r.Time.Local().Format("03:04PM"))

		default:
			result := strings.SplitN(part, ":", 2)
			if len(result) != 2 && result[0] != "attr" {
				// some other string so just print it as-is
				fmt.Fprintf(buf, "%s", part)
				continue partsloop
			}
			attrKey := result[1]

			_, attrs := h.resolveRecordAttrs(ctx, buf, r, false)
			if v, ok := attrs[attrKey]; ok {
				h.printAttr(ctx, r, buf, attrKey, v, printedAttrs)
			}
		}
	}

	// finally - write the message
	buf.WriteByte('\n')
	_, err := h.options.Writer.Write(buf.Bytes())
	return err
}

// WithAttrs creates a new handler from the existing one adding the given attributes to it.
func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		options: h.options,
		attrs:   appendAttrsToGroup(h.groups, h.attrs, attrs),
		groups:  h.groups,
	}
}

// WithGroup creates a new handler from the existing one adding the given group to it.
func (h Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		options: h.options,
		attrs:   h.attrs,
		groups:  append(h.groups, name),
	}
}

// colorize handles colorizing the given string if colorization is enabled.
func (h Handler) colorize(buf *bytes.Buffer, c *color.Color, a interface{}) {
	if h.options.NoColor {
		fmt.Fprintf(buf, "%v", a)
	} else {
		c.Fprintf(buf, "%v", a)
	}
}

// printAttr handles printing the attribute to the buffer.
//
// Be sure to call resolveAttrs() before calling this function.
func (h Handler) printAttr(ctx context.Context, r slog.Record, buf *bytes.Buffer, attrKey string,
	attrValue slog.Value, printedAttrs map[string]bool) {

	// already printed the given key
	if _, ok := printedAttrs[attrKey]; ok {
		return
	}

	// use the attribute-specific formatter (attr:NAME)
	formatterKey := fmt.Sprintf(AttrKey, attrKey)
	if fn, ok := h.options.PartsFormatter[formatterKey]; ok && fn != nil {
		fmt.Fprintf(buf, "%s", fn(ctx, formatterKey, r))
		printedAttrs[attrKey] = true
		return
	}

	// use the general attribute formatter
	if fn, ok := h.options.PartsFormatter[AttrsKey]; ok && fn != nil {
		fmt.Fprintf(buf, "%s", fn(ctx, formatterKey, r))
		printedAttrs[attrKey] = true
		return
	}

	// use the default formatter
	h.colorize(buf, color.New(color.FgHiBlue), attrKey)
	fmt.Fprintf(buf, "=%v", attrValue.Any())
	printedAttrs[attrKey] = true
}

// printAttrs prints the attributes to the buffer.
//
// Be sure to call resolveAttrs() before calling this function.
func (h Handler) printAttrs(ctx context.Context, r slog.Record, buf *bytes.Buffer, keyOrder []string,
	attrs map[string]slog.Value, printedAttrs map[string]bool) {

	lastBufLen := buf.Len()
	for _, attrKey := range keyOrder {
		// already printed the given key
		if _, ok := printedAttrs[attrKey]; ok {
			continue
		}

		// only print the parts separator if we actually printed something before
		currentBufLen := buf.Len()
		if currentBufLen > 0 && currentBufLen != lastBufLen {
			fmt.Fprintf(buf, "%s", h.options.PartsSeparator)
			currentBufLen = buf.Len()
		}
		lastBufLen = currentBufLen

		// print the attribute
		h.printAttr(ctx, r, buf, attrKey, attrs[attrKey], printedAttrs)
	}
}

// resolveAttrs resolves all group attributes and "flattens" them out.
//
// This function should only be called from resolveRecordAttrs().
func (h Handler) resolveAttrs(ctx context.Context, buf *bytes.Buffer, attrs []slog.Attr, keyPrefix string) (
	[]string, map[string]slog.Value) {

	resolved := map[string]slog.Value{}
	attrKeys := []string{}
	for _, attr := range attrs {
		// resolve all group attributes
		if attr.Value.Kind() == slog.KindGroup {
			groupKeys, groupAttr := h.resolveAttrs(ctx, buf, attr.Value.Group(), attr.Key)
			attrKeys = append(attrKeys, groupKeys...)
			for k, v := range groupAttr {
				resolved[k] = v
			}
			continue
		}

		// add keys and values
		attrKey := attr.Key
		if keyPrefix != "" {
			attrKey = fmt.Sprintf("%s.%s", keyPrefix, attrKey)
		}
		attrKeys = append(attrKeys, attrKey)
		resolved[attrKey] = attr.Value
	}
	return attrKeys, resolved
}

// resolveRecordAttrs resolves all attributes and group attributes and "flattens" everything out.
func (h Handler) resolveRecordAttrs(ctx context.Context, buf *bytes.Buffer, r slog.Record, sortKeys bool) (
	[]string, map[string]slog.Value) {

	attrs := []slog.Attr{}
	r.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})

	keys, resolvedAttrs := h.resolveAttrs(ctx, buf, attrs, "")
	if sortKeys {
		sort.Strings(keys)
	}
	return keys, resolvedAttrs
}
