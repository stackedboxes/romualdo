# Implementation

## Error handling strategy

Simple, just a little help from the type system to ensure we exit with the
proper status:

* Exported functions on all packages shall only return `err.Error`s.
    * These know their proper exit code.
    * And also allow testing code to know what value would be returned to the
      shell just by looking at the error.
* Commands use the `reportAndExitOnError()` and `reportAndExit()` helpers to
  handle error reporting.
