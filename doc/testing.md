# Testing

Notes on the format expected by `romualdo dev test`.

**Warning.** Right now, these are more musings on what to make than a
description of the actual implementation (which is extremely bare and not very
much like what is described here).

## Concepts

A **test case** is declaratively defined in a `test.toml` file. Each test case
is comprised of a sequence of **steps** (though many test cases will have one
single step, and the syntax is simplified for this case).

The term **test suite** is loosely used to describe a set of test cases. It is
just a directory where we recursively look for `test.toml` files.

### Single-step versus multi-step test cases

Let's start with the simpler case, which is a single-step test case. It would be
defined like this:

```toml
# test.toml for a single-step test case

key1 = "value"
key2 = 123
key3 = [
    "one",
    "two",
    "three",
]
```

The keys and values here are mde up, but the structure is real: a single-step
test case is just a bunch of key/values at the top-level of the `test.toml`
file.

A multi-step test case is not very much unlike a single-step one. In TOML terms,
it contains an array of tables called `step`, in which each member is exactly
like in the single-step case. Like this:

```toml
# test.toml for a multi-step test case

# Here's the first step
[[step]]
    key1 = "value"
    key2 = 123
    key3 = [
        "one",
        "two",
        "three",
    ]

# And here's the second one
[[step]]
    key1 = "another value"
    key2 = 123
    key3 = [
        "1",
        "22",
        "333",
    ]
```

There's another handy possibility with multi-step test cases: any keys you
define at top-level are used as defaults for the actual steps. The top-level
keys do not define a step, they are just a handy way to share values between
steps.

```toml
# test.toml for a multi-step test case with shared defaults

# These are just default values, not a real step.
key2 = 123
key3 = [
    "A",
    "B",
    "C",
]

# This is the first step. It "inherits" the key2 and key3 values from the
# top-level declarations above.
[[step]]
    key1 = "value"

# And this is the second step. It "inherits" key3 from the top-level
# declarations, but overrides key2.
[[step]]
    key1 = "another value"
    key2 = 321
```

## Reference

These are the keys recognized by `romualdo dev test`.

### `type`

*Default: `build-and-run`.*

Steps can be of different types, depending on what you want the step to do. So,
in a sense this is the most important key, as it determines even which other
keys are relevant.

Possible values are:

* `build`: The step builds the source code. It can check for error codes and
  error messages.
* `run`: The step runs the Storyworld built on a previous step. It can send
  input, and check if it gets the expected output.
* `build-and-run`: Equivalent to `build` followed by `run`; the `run` part is
  skipped if the build fails.
* `save-state`: The step saves the VM state.
* `load-state`: The step loads the VM state (assumed to have been previously
  saved).

### `sourceDir`

*Valid for:* `build`, `build-and-run`.  
*Default:* `src`.

Defines the directory where the Storyworld source code will be looked for. This
is relative to the directory where `test.toml` is.

### `input`

*Valid for:* `run`, `build-and-run`.  
*Default:* `[]`

An array of strings, which will be send as input to the Storyworld. Each element
in the array will be sent at a time, for each time the Storyworld `listen`s.

It is expected that all elements in the array will be used when the execution
completes.

### `output`

*Valid for:* `run`, `build-and-run`.  
*Default:* `[]`

An array of strings, which represent the expected output from the Storyworld.

TODO: I initially wrote the following sentence, but it's wrong and need to be
better thought out. "`input` and `output` are used in lockstep: first an output
is used, then an input is send, then a new output is taken, and so on. So, there
must be one output more than inputs."

### `exitCode`

*Valid for:* All `type`s.  
*Default:* `0`

This is the expected exit code of the `romualdo` tool when running the test
case.

It is relevant only for the last step, as all previous steps are expected to be
successful.

### `errorMessages`

*Valid for:* All `type`s.  
*Default:* `[]`

This is an array of strings, each of which representing an error message
expected to be present in the output. Each string is really interpreted as a
regular expression that must match the standard error.

It is relevant only for the last step, as all previous steps are expected to be
successful.
