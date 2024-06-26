# Changelog

## Unreleased

No unreleased changes

## v0.6.3 (Released 2024-04-01)

* Added `LoggingService` interface to more easily swap loggers in external libraries
* Removed `slogx`-specific context functions
* Added `LoggingService` context functions

## v0.6.2 (Released 2024-03-30)

* Added `ParseLevel()` function
* Updated `condition` object to be exported as `Condition`
* Added `SetDefault()` function
  
## v0.6.1 (Released 2024-02-12)

* Updated `LevelMin` to be `-2147483647` and `LevelUnknown` to be `-2147483648`
* Updated `LevelMax` to be `2147483646` and `LevelDisabled` to be `2147483647`
* Fixed missing `Level()` function in `HTTPHandler`
* Fixed `Level` variable in `HTTPHandlerOptions`
* Removed `nilHandler` and replaced with `Nil()` logger

## v0.6.0 (Released 2024-01-27)

* Added `LevelUnknown` level which can be used in cases where a level cannot be determined
* Replaced `LevelNone` with `LevelDisabled` and fixed value so no messages will be logged
* Added `LevelMin` and `LevelMax` which corresponding to respective minimum and maximum level values
* Added `NilHandler` to simply discard all messages (as an alternative to setting a handler to `LevelNone`)
* Updated built-in handlers to use `LevelVar` for level so they can be changed dynamically
* Added `LevelVarHandler` interface for more easily accessing a handler's dynamic level stored in its options
* Refactored formatter and handler option context functions
* Added capabilities to use regular expressions for console formatter parts
* Fixed bug in `Enable` and `Handle` functions with conditional, failover, multi, pipe and round robin handlers where messages below log level threshold would still be logged
* Added `HTTPHandler` for sending messages via HTTP POST requests
* Added ability to store logger and handlers in context by name in order to store multiple loggers and/or handlers within the same context
* Added functions to be able to retrieve logger from context based on name or the logger marked as "active" in the context
* Added functions to store and retrieve error attribute name for log messages from context
  
## v0.3.2 (Released 2023-10-06)

* Added `LevelNone` level which can be used an option to not log any messages
  
## v0.3.1 (Released 2023-10-02)

* Updated requirements to use `go` version `1.21` or later in order to remove experimental library
* Added `Default()` function to return default logger

## v0.2.0 (Released 2023-09-25)

* Added `ErrorRecord` type
* Added `LogRecord()` function to `Logger` in order to more easily log `ErrorRecord` objects
* Updated formatters and handlers to use consistent methods for setting/getting options context
  
## v0.1.0 (Released 2023-08-29)

* Initial release of the module
