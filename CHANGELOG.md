# Changelog

## Unreleased

No unreleased changes

## v0.4.0 (Released 2023-10-07)

* Added `LevelUnknown` level which can be used in cases where a level cannot be determined
* Added `DynamicLevelHandler` interface in order to be able to more easily change levels for an existing handler
* Added `NilHandler` to simply discard all messages (as an alternative to setting a handler to `LevelNone`)
* Added `DynamicLevelHandler` support to built-in handlers
  
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
