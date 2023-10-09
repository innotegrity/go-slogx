# Changelog

## Unreleased

No unreleased changes

## v0.5.3 (Released 2023-xx-xx)

* Added capabilities to use regular expressions for console formatter parts
* Fixed bug in `Enable` and `Handle` functions with conditional, failover, multi, pipe and round robin handlers where messages below log level threshold would still be logged

## v0.5.2 (Released 2023-10-08)

* Added `LevelUnknown` level which can be used in cases where a level cannot be determined
* Added `NilHandler` to simply discard all messages (as an alternative to setting a handler to `LevelNone`)
* Changed built-in handlers to use `LevelVar` for level so they can be changed dynamically
* Added ability to store logger and handlers in context by name in order to store multiple loggers and/or handlers within the same context
* Added `LevelVarHandler` interface for more easily accessing a handler's dynamic level stored in its options
* Refactored formatter and handler option context functions
  
## v0.3.2 (Released 2023-10-06)

* Added `LevelNone` level which can be used an option to not log any messages
  
## v0.3.1 (Released 2023-10-02)

* Moved to requiring `go` version `1.21` or later in order to remove experimental library
* Added `Default()` function to return default logger

## v0.2.0 (Released 2023-09-25)

* Added `ErrorRecord` type
* Added `LogRecord()` function to `Logger` in order to more easily log `ErrorRecord` objects
* Updated formatters and handlers to use consistent methods for setting/getting options context
  
## v0.1.0 (Released 2023-08-29)

* Initial release of the module
