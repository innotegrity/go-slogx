package handler

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"log/slog"

	"go.innotegrity.dev/generic"
	"go.innotegrity.dev/slogx"
	"go.innotegrity.dev/slogx/formatter"
)

// fileHandlerOptionsContext can be used to retrieve the options used by the handler from the context.
type fileHandlerOptionsContext struct{}

// FileHandlerOptions holds options for the file handler.
type FileHandlerOptions struct {
	// DirMode is the mode to use when creating directories.
	//
	// By default, directories will be created with mode 0755.
	DirMode fs.FileMode

	// Filename is the name of the log file to write to.
	//
	// This is a required option.
	Filename string

	// FileMode is the mode to use when creating log files.
	//
	// By default, files will be created with mode 0640.
	FileMode fs.FileMode

	// Level is the minimum log level to write to the handler.
	Level slogx.Level

	// MaxFileCount indicates the maximum number of log files to keep, including the active log file.
	//
	// By default, this is set to 5. If this value is negative, an unlimited number of files will be kept.
	MaxFileCount int

	// MaxFileSize indicates the maximum size of any log file, in bytes. Once a file reaches this size,
	// it will be rotated automatically.
	//
	// By default, the maximum file size will be 10MB (10000000 bytes). If this value is negative, the file will
	// never be rotated.
	MaxFileSize int64

	// RecordFormatter specifies the formatter to use to format the record before sending it to Slack.
	//
	// If no formatter is supplied, formatter.DefaultJSONFormatter is used to format the output.
	RecordFormatter formatter.BufferFormatter
}

// DefaultFileHandlerOptions returns a default set of options for the handler.
func DefaultFileHandlerOptions() FileHandlerOptions {
	return FileHandlerOptions{
		DirMode:         0755,
		FileMode:        0640,
		Level:           slogx.LevelInfo,
		MaxFileCount:    5,
		MaxFileSize:     10000000,
		RecordFormatter: formatter.DefaultJSONFormatter(),
	}
}

// GetFileHandlerOptionsFromContext retrieves the options from the context.
//
// If the options are not set in the context, a set of default options is returned instead.
func GetFileHandlerOptionsFromContext(ctx context.Context) *FileHandlerOptions {
	o := ctx.Value(fileHandlerOptionsContext{})
	if o != nil {
		if opts, ok := o.(*FileHandlerOptions); ok {
			return opts
		}
	}
	opts := DefaultFileHandlerOptions()
	return &opts
}

// AddToContext adds the options to the given context and returns the new context.
func (o *FileHandlerOptions) AddToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, fileHandlerOptionsContext{}, o)
}

// fileHandler is a log handler that writes records to a file.
type fileHandler struct {
	activeGroup     string
	attrs           []slog.Attr
	currentFileSize *int64
	file            *os.File
	groups          []string
	options         FileHandlerOptions
	writeLock       *sync.Mutex
}

// NewFileHandler creates a new handler object.
func NewFileHandler(opts FileHandlerOptions) (*fileHandler, error) {
	// validate required options
	if opts.Filename == "" {
		return nil, errors.New("filename is required and cannot be empty")
	}

	// set default options
	if opts.DirMode == 0 {
		opts.DirMode = 0755
	}
	if opts.FileMode == 0 {
		opts.FileMode = 0640
	}
	if opts.MaxFileCount == 0 {
		opts.MaxFileCount = 5
	}
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = 10000000
	}

	// create the handler
	currentFileSize := int64(0)
	return &fileHandler{
		attrs:           []slog.Attr{},
		currentFileSize: &currentFileSize,
		groups:          []string{},
		options:         opts,
		writeLock:       &sync.Mutex{},
	}, nil
}

// Enabled determines whether or not the given level is enabled in this handler.
func (h fileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.options.Level.Level()
}

// Handle actually handles writing the record to the file.
//
// Any attributes duplicated between the handler and record, including within groups, are automaticlaly removed.
// If a duplicate is encountered, the last value found will be used for the attribute's value.
func (h *fileHandler) Handle(ctx context.Context, r slog.Record) error {
	handlerCtx := h.options.AddToContext(ctx)
	attrs := slogx.ConsolidateAttrs(h.attrs, h.activeGroup, r)

	// format the output into a buffer
	var buf *slogx.Buffer
	var err error
	if h.options.RecordFormatter != nil {
		buf, err = h.options.RecordFormatter.FormatRecord(handlerCtx, r.Time, slogx.Level(r.Level), r.PC, r.Message,
			attrs)
	} else {
		f := formatter.DefaultJSONFormatter()
		buf, err = f.FormatRecord(handlerCtx, r.Time, slogx.Level(r.Level), r.PC, r.Message, attrs)
	}
	if err != nil {
		return err
	}

	// write the buffer to the file
	return h.write(buf)
}

// Level returns the current logging level that is in use by the handler.
func (h fileHandler) Level() slogx.Level {
	return h.options.Level
}

// Shutdown is responsible for cleaning up resources used by the handler.
func (h fileHandler) Shutdown(continueOnError bool) error {
	if h.file != nil {
		h.file.Close()
	}
	return nil
}

// WithAttrs creates a new handler from the existing one adding the given attributes to it.
func (h fileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &fileHandler{
		attrs:           h.attrs,
		currentFileSize: h.currentFileSize,
		file:            h.file,
		groups:          h.groups,
		options:         h.options,
		writeLock:       h.writeLock,
	}
	if h.activeGroup == "" {
		newHandler.attrs = append(newHandler.attrs, attrs...)
	} else {
		newHandler.attrs = append(newHandler.attrs, slog.Group(h.activeGroup, generic.AnySlice(attrs)...))
		newHandler.activeGroup = h.activeGroup
	}
	return newHandler
}

// WithGroup creates a new handler from the existing one adding the given group to it.
func (h fileHandler) WithGroup(name string) slog.Handler {
	newHandler := &fileHandler{
		attrs:           h.attrs,
		currentFileSize: h.currentFileSize,
		file:            h.file,
		groups:          h.groups,
		options:         h.options,
		writeLock:       h.writeLock,
	}
	if name != "" {
		newHandler.groups = append(newHandler.groups, name)
		newHandler.activeGroup = name
	}
	return newHandler
}

// WithLevel returns a new handler with the given logging level set.
func (h fileHandler) WithLevel(level slogx.Level) slogx.DynamicLevelHandler {
	options := h.options
	options.Level = level
	return &fileHandler{
		activeGroup:     h.activeGroup,
		attrs:           h.attrs,
		currentFileSize: h.currentFileSize,
		file:            h.file,
		groups:          h.groups,
		options:         options,
		writeLock:       h.writeLock,
	}
}

// openFile opens the log file for writing or creates it and any parent folders if they do not exist.
func (h *fileHandler) openFile() error {
	// make sure parent folder exists
	dir := filepath.Dir(h.options.Filename)
	_, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		err := os.MkdirAll(dir, h.options.DirMode)
		if err != nil {
			return err
		}
	}

	// open the file for creation/appending
	file, err := os.OpenFile(h.options.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, h.options.FileMode)
	if err != nil {
		return err
	}

	// get the current file size
	info, err := os.Stat(h.options.Filename)
	if err != nil {
		return err
	}

	// save the file handle and size
	*h.currentFileSize = info.Size()
	h.file = file
	return nil
}

// rotateFiles rotates the current log file and existing log files and opens a new file for writing.
func (h *fileHandler) rotateFiles() error {
	// close existing log file
	if h.file != nil {
		h.file.Close()
	}

	// rotate previous files
	dir, file := filepath.Split(h.options.Filename)
	ext := filepath.Ext(file)
	file = file[:len(file)-len(ext)]
	for i := h.options.MaxFileCount - 1; i > 0; i-- {
		src := filepath.Join(dir, fmt.Sprintf("%s_%d%s", file, i, ext))
		_, err := os.Stat(src)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}

		// remove the file
		if i == (h.options.MaxFileCount - 1) {
			if err := os.Remove(src); err != nil {
				return err
			}
			continue
		}

		// rotate the file
		dest := filepath.Join(dir, fmt.Sprintf("%s_%d%s", file, i+1, ext))
		if err := os.Rename(src, dest); err != nil {
			return err
		}
	}

	// rotate the current file
	dest := filepath.Join(dir, fmt.Sprintf("%s_1%s", file, ext))
	if err := os.Rename(h.options.Filename, dest); err != nil {
		return err
	}

	// open the new log file
	return h.openFile()
}

// write handles writing the buffer contents to the file.
func (h *fileHandler) write(buf *slogx.Buffer) error {
	h.writeLock.Lock()
	defer h.writeLock.Unlock()

	// open the file if it's not already open
	if h.file == nil {
		if err := h.openFile(); err != nil {
			return err
		}
	}

	// rotate logs if message will cause the file to exceed the maximum desired size
	if (*h.currentFileSize + int64(buf.Len())) > h.options.MaxFileSize {
		if err := h.rotateFiles(); err != nil {
			return err
		}
	}

	// write message to file
	bytesWritten, err := h.file.Write(buf.Bytes())
	*h.currentFileSize += int64(bytesWritten)
	return err
}
