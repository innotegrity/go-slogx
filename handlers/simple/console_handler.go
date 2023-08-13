package simple

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/mattn/go-colorable"
	"go.innotegrity.dev/slogx"
	"go.innotegrity.dev/slogx/internal/buffer"
	"go.innotegrity.dev/slogx/internal/utils"
	"go.innotegrity.dev/slogx/middleware"
	"golang.org/x/exp/slog"
)

const (
	// AttrKey is the key used for PartsOrder for a specific attribute. You must supply the attribute name via
	// fmt.Sprintf(AttrKey, ATTRIBUTE_NAME) where ATTRIBUTE_NAME is the actual attribute.
	AttrPart = "attr:%s"

	// AttrsKey is the key used for PartsOrder for all attributes in no particular order.
	AllAttrsPart = "attrs"

	// AllAttrsAlphaPart is the key used for PartsOrder for all attributes in alphabetical order.
	AllAttrsAlphaPart = "attrs_alpha"

	// LevelPart is the key used for PartsOrder for the message level.
	LevelPart = "level"

	// MessagePart is the key used for PartsOrder for the message itself.
	MessagePart = "message"

	// SourcePart is the key used for PartsOrder for the source code location for the record.
	SourcePart = "source"

	// TimePart is the key used for PartsOrder for the time of the message/record.
	TimePart = "time"
)

// ConsoleHandlerOptionsContext can be used to retrieve the options used by the console handler from the context
// passed to a formatter function.
type ConsoleHandlerOptionsContext struct{}

// ConsoleHandlerOptions holds the options for the console handler.
type ConsoleHandlerOptions struct {
	// AttrFormatter is the middleware formatting function to call to format any attribute.
	//
	// Attribute values are always resolved before and after calling the function.
	//
	// If nil, attributes are simply printed unchanged as key=value.
	AttrFormatter middleware.FormatAttrFn

	// ContinueOnError determines whether or not to continue if an error occurs running middleware.
	ContinueOnError bool

	// EnableColor determines whether or not to enable colorized output.
	EnableColor bool

	// Level is the minimum log level to write to the handler.
	//
	// By default, the level will be set to slog.LevelInfo if not supplied.
	Level slog.Leveler

	// IgnoreAttrs is a list of attributes to ignore and not print.
	IgnoreAttrs []string

	// LevelFormatter is the middleware formatting function to call to format the level.
	//
	// If nil, the level is printed using middleware.FormatLevelValueDefault().
	LevelFormatter middleware.FormatLevelValueFn

	// MessageFormatter is the middlware formatting function to call to format the message.
	//
	// If nil, the message is printed as-is.
	MessageFormatter middleware.FormatMessageValueFn

	// PartOrder is the order in which to print the various parts of the message.
	//
	// The following values are valid for the string:
	// AttrPart - specific attribute - use Sprintf(AttrKey, NAME) where NAME is the specific attribute to use; use
	//            a single period (.) to separate group name from attribute name if the attribute is nested within a
	//            group
	// AllAttrsPart - all attributes in any order
	// AllAttrsAlphaPart - all attributes in alphabetical order
	// LevelPart - the log level of the message from the record as a string
	// MessagePart - the message from the record
	// SourcePart - the location at which the message was logged
	// TimePart - the time of the message from the record
	//
	// If any other string is specified other than those above, it is simply printed as-is.
	//
	// Any attribute will only be output once.  If a specific attribute is output using the AttrPart, it will not
	// be repeated later if AttrsPart or AllAttrsAlphaPart is specified. Likewise, AllAttrsPart and AllAttrsAlphaPart
	// should never both be specified as only the first one will be used.
	//
	// If an attribute is specified but not present in the record, it is ignored.
	//
	// By default the format will be "TimePart LevelPart SourcePart > MessagePart AllAttrsAlphaPart".
	PartOrder []string

	// PartSeparator is how to separate the parts when they're being printed. Must be at least 1 character.
	//
	// By default, a space is used.
	PartSeparator string

	// SourceFormatter is the middleware formatting function to call to format the source code location where the record
	// was created.
	//
	// If nil, the source code location is printed using middleware.FormatSourceValueDefault().
	SourceFormatter middleware.FormatSourceValueFn

	// SpecificAttrFormatter is the middleware formatting function to call to format a specific attribute.
	//
	// Attribute values are always resolved before and after calling the function.
	//
	// If nil or if the attribute does not exist in the map, the default is to fall back to the AttrFormatter function.
	SpecificAttrFormatter map[string]middleware.FormatAttrFn

	// TimeFormatter is the middleware formatting function to call to the time of the record.
	//
	// If nil, the time is printed using middleware.FormatTimeValueDefault().
	TimeFormatter middleware.FormatTimeValueFn

	// Writer is where to write the output to.
	//
	// By default, messages are written to os.Stdout if not supplied.
	Writer io.Writer
}

// consoleHandler is a log handler that writes records to an io.Writer, typically a console in a specified format.
type consoleHandler struct {
	attrs       []slog.Attr
	groups      []string
	ignoredAttr map[string]bool
	options     ConsoleHandlerOptions
	writeLock   *sync.Mutex
}

// NewConsoleHandler creates a new handler object.
func NewConsoleHandler(opts ConsoleHandlerOptions) *consoleHandler {
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}
	if len(opts.PartOrder) == 0 {
		opts.PartOrder = []string{
			TimePart,
			LevelPart,
			SourcePart,
			">",
			MessagePart,
			AllAttrsAlphaPart,
		}
	}
	if opts.PartSeparator == "" {
		opts.PartSeparator = " "
	}
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	if (opts.Writer == os.Stdout || opts.Writer == os.Stderr) && opts.EnableColor {
		opts.Writer = colorable.NewColorable(opts.Writer.(*os.File))
	}
	ignoredAttr := map[string]bool{}
	for _, attr := range opts.IgnoreAttrs {
		ignoredAttr[attr] = true
	}
	return &consoleHandler{
		attrs:       []slog.Attr{},
		groups:      []string{},
		ignoredAttr: ignoredAttr,
		options:     opts,
		writeLock:   &sync.Mutex{},
	}
}

// Enabled determines whether or not the given level is enabled in this handler.
func (h consoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.options.Level.Level()
}

// Handle actually handles writing the record to the output writer.
func (h *consoleHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := buffer.New()
	printedAttrs := map[string]bool{}
	lastBufLen := 0
	record := r.Clone()

	var val slog.Value
	var err error
	handlerCtx := context.WithValue(ctx, ConsoleHandlerOptionsContext{}, &h.options)
	for _, part := range h.options.PartOrder {
		// only print the parts separator if we actually printed something before
		currentBufLen := buf.Len()
		if currentBufLen > 0 && currentBufLen != lastBufLen {
			fmt.Fprintf(buf, "%s", h.options.PartSeparator)
			currentBufLen = buf.Len()
		}
		lastBufLen = currentBufLen

		switch part {
		case AllAttrsAlphaPart:
			keyOrder, attrs := utils.FlattenRecordAttrs(handlerCtx, record, true)
			if err = h.printAttrs(handlerCtx, buf, keyOrder, attrs, printedAttrs); err != nil &&
				!h.options.ContinueOnError {
				return err
			}

		case AllAttrsPart:
			keyOrder, attrs := utils.FlattenRecordAttrs(handlerCtx, record, false)
			if err = h.printAttrs(handlerCtx, buf, keyOrder, attrs, printedAttrs); err != nil &&
				!h.options.ContinueOnError {
				return err
			}

		case LevelPart:
			if h.options.LevelFormatter != nil {
				val, err = h.options.LevelFormatter(handlerCtx, slogx.Level(record.Level))
			} else {
				val, err = middleware.FormatLevelValueDefault(handlerCtx, slogx.Level(record.Level))
			}
			if err != nil && !h.options.ContinueOnError {
				return err
			}
			fmt.Fprintf(buf, "%s", val.Resolve().String())

		case MessagePart:
			if h.options.MessageFormatter != nil {
				val, err = h.options.MessageFormatter(handlerCtx, record.Message)
			} else {
				val = slog.StringValue(record.Message)
				err = nil
			}
			if err != nil && !h.options.ContinueOnError {
				return err
			}
			fmt.Fprintf(buf, "%s", val.Resolve().String())

		case SourcePart:
			if h.options.SourceFormatter != nil {
				val, err = h.options.SourceFormatter(handlerCtx, record.PC)
			} else {
				val, err = middleware.FormatSourceValueDefault(handlerCtx, record.PC)
			}
			if err != nil && !h.options.ContinueOnError {
				return err
			}
			fmt.Fprintf(buf, "%s", val.Resolve().String())

		case TimePart:
			if h.options.TimeFormatter != nil {
				val, err = h.options.TimeFormatter(handlerCtx, record.Time)
			} else {
				val, err = middleware.FormatTimeValueDefault(handlerCtx, record.Time)
			}
			if err != nil && !h.options.ContinueOnError {
				return err
			}
			fmt.Fprintf(buf, "%s", val.Resolve().String())

		default:
			result := strings.SplitN(part, ":", 2)
			if len(result) != 2 && result[0] != "attr" {
				// raw string without formatting
				fmt.Fprint(buf, part)
			} else {
				attrKey := result[1]
				_, attrs := utils.FlattenRecordAttrs(handlerCtx, record, false)
				if v, ok := attrs[attrKey]; ok {
					if err = h.printAttr(handlerCtx, buf, attrKey, v, printedAttrs); err != nil && !h.options.ContinueOnError {
						return err
					}
				}
			}
		}
	}

	// finally - write the message
	buf.WriteByte('\n')
	h.writeLock.Lock()
	defer h.writeLock.Unlock()
	_, err = h.options.Writer.Write(buf.Bytes())
	return err
}

// Shutdown is responsible for cleaning up resources used by the handler.
func (h consoleHandler) Shutdown(continueOnError bool) error {
	return nil
}

// WithAttrs creates a new handler from the existing one adding the given attributes to it.
func (h consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &consoleHandler{
		options:   h.options,
		attrs:     append(h.attrs, attrs...),
		groups:    h.groups,
		writeLock: h.writeLock,
	}
}

// WithGroup creates a new handler from the existing one adding the given group to it.
func (h consoleHandler) WithGroup(name string) slog.Handler {
	return &consoleHandler{
		options:   h.options,
		attrs:     h.attrs,
		groups:    append(h.groups, name),
		writeLock: h.writeLock,
	}
}

// printAttr handles printing the attribute to the buffer.
func (h consoleHandler) printAttr(ctx context.Context, buf *buffer.Buffer, attrKey string, attrValue slog.Value,
	printedAttrs map[string]bool) error {

	// ignore the given key
	if _, ok := h.ignoredAttr[attrKey]; ok {
		return nil
	}

	// already printed the given key
	if _, ok := printedAttrs[attrKey]; ok {
		return nil
	}

	// print the attribute
	var err error
	formattedKey := attrKey
	formattedValue := attrValue
	if fn, ok := h.options.SpecificAttrFormatter[attrKey]; ok && fn != nil {
		formattedKey, formattedValue, err = fn(ctx, formattedKey, formattedValue.Resolve())
		if err != nil && !h.options.ContinueOnError {
			return err
		}
	} else if h.options.AttrFormatter != nil {
		formattedKey, formattedValue, err = h.options.AttrFormatter(ctx, formattedKey, formattedValue.Resolve())
		if err != nil && !h.options.ContinueOnError {
			return err
		}
	}
	fmt.Fprintf(buf, "%s=%v", formattedKey, formattedValue.Resolve().Any())
	printedAttrs[attrKey] = true
	return nil
}

// printAttrs prints the attributes to the buffer.
func (h consoleHandler) printAttrs(ctx context.Context, buf *buffer.Buffer, keyOrder []string,
	attrs map[string]slog.Value, printedAttrs map[string]bool) error {

	lastBufLen := buf.Len()
	for _, attrKey := range keyOrder {
		// already printed the given key
		if _, ok := printedAttrs[attrKey]; ok {
			continue
		}

		// only print the parts separator if we actually printed something before
		currentBufLen := buf.Len()
		if currentBufLen > 0 && currentBufLen != lastBufLen {
			fmt.Fprintf(buf, "%s", h.options.PartSeparator)
			currentBufLen = buf.Len()
		}
		lastBufLen = currentBufLen

		// print the attribute
		if err := h.printAttr(ctx, buf, attrKey, attrs[attrKey], printedAttrs); err != nil && !h.options.ContinueOnError {
			return err
		}
	}
	return nil
}
