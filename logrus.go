package slogx

/*
// InitLogrusInterceptor initializes the logrus global logger with an expected format so that they can be
// intercepted and parsed by ParseLogrusMessages.
func (l *Logger) InterceptLogrusMessages(writer *io.PipeWriter) {
	logrus.SetLevel(loggerLevelToLogrusLevel(l.level))
	logrus.SetOutput(writer)
	logrus.SetFormatter(&logrus.JSONFormatter{
		DataKey:          "data",
		DisableTimestamp: true,
	})
	if l.IsDebugEnabled() {
		logrus.SetReportCaller(true)
	}
}

// ParseLogrusMessageErrorHandler is called when an error occurs while parsing a logrus message.
type ParseLogrusMessageErrorHandler func(error, []byte)

// ParseLogrusMessages uses a reader paired with the output writer from InitLogrusInterceptor to read messages
// from the pipe in a blocking manner until the writer is closed.
func (l *Logger) ParseLogrusMessages(reader *io.PipeReader, handler ParseLogrusMessageErrorHandler) {
	buf := bufio.NewReader(reader)
	self := l

	for {
		line, _, err := buf.ReadLine()
		if err == nil || err == io.EOF {
			if len(line) > 0 {
				var msg logrusMessage
				if unmarshalErr := json.Unmarshal([]byte(line), &msg); unmarshalErr != nil {
					if handler != nil {
						handler(err, line)
					}
				} else {
					level, _ := logrus.ParseLevel(msg.Level)
					loggerLevel := logrusLevelToLoggerLevel(level)
					messageLogger := self.WithLevel(loggerLevel)
					if msg.Method != "" {
						messageLogger = messageLogger.Str("method", msg.Method)
					}
					for k, v := range msg.Data {
						messageLogger = messageLogger.Interface(k, v)
					}
					messageLogger.Msg(msg.Message)
				}
			}
			if err == io.EOF {
				break
			}
		} else {
			if handler != nil {
				handler(err, nil)
			}
		}
	}
}

// logrusMessage represents a JSON-formatted log message.
type logrusMessage struct {
	Level   string                 `json:"level"`
	Message string                 `json:"msg"`
	Method  string                 `json:"func"`
	Data    map[string]interface{} `json:"data"`
}

// loggerLevelToLoggerLevel converts our logger logging levels to logrus logging levels.
func loggerLevelToLogrusLevel(l Level) logrus.Level {
	switch l {
	case PanicLevel:
		return logrus.PanicLevel
	case FatalLevel:
		return logrus.FatalLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	case WarnLevel:
		return logrus.WarnLevel
	case InfoLevel:
		return logrus.InfoLevel
	case DebugLevel:
		return logrus.DebugLevel
	case TraceLevel:
		return logrus.TraceLevel
	}
	return 0
}

// logrusLevelToLoggerLevel converts logrus logging levels to our logger logging levels.
func logrusLevelToLoggerLevel(l logrus.Level) Level {
	switch l {
	case logrus.PanicLevel:
		return PanicLevel
	case logrus.FatalLevel:
		return FatalLevel
	case logrus.ErrorLevel:
		return ErrorLevel
	case logrus.WarnLevel:
		return WarnLevel
	case logrus.InfoLevel:
		return InfoLevel
	case logrus.DebugLevel:
		return DebugLevel
	case logrus.TraceLevel:
		return TraceLevel
	}
	return NoLevel
}
*/
