# File Formats

## Compiled Storyworld

All integers and floating-point numbers are stored in little endian byte order.
For signed integers, two's complement is used.

### Compiled Storyworld Header

* An 8-byte "magic number" comprised of the string `RmldCSW` followed by a SUB
  character (`0x1A`, which in times long gone used to represent a "soft
  end-of-file"). These are written to the file in this exact order, i.e., the
  first byte on the file is `R`, the second is `m`, and so on.
* A `uint32` with the version (currently 0).

### Compiled Storyworld Payload

#### Releases

* A `uint32` with the number of releases.
* For each of the releases:
    * An `uint32` with the length of the string in bytes.
    * An array of bytes, with the string data encoded in UTF-8.

#### Constants

* A `uint32` with the number of constants.
* Each of the constants, as Values (see below).

#### Chunks

* A `uint32` with the number of Chunks.
* Each of the Chunks, which looks like this:
    * A `uint32` with Chunk size.
    * An array of bytes, with the bytecode. The opcodes and instruction format
      are documented in [Instruction Set](instruction_set.md).
    * A Boolean telling whether this Chunk is part of a release or not. Encoded
      as a byte, either 0 (false) or 1 (true).
    * The code hash of the Procedure that was compiled into this Chunk. This is
      an array of 32 bytes.

#### Procedure Data

* A `uint32` with the number of Procedures. (This doesn't count take versions
  into account: a Procedure with multiple version counts as just one here.)
* Each of the Procedures, which looks like this:
    * The fully-qualified Procedure name, as a string (`uint32` length, followed
      by UTF-8 data).
    * The return type of the Procedure, [encoded as a type](#types).
    * A `uint32` with the number of arguments the Procedure has.
    * For each argument:
        * The argument [type](#types).
    * A `uint32` with the number of versions the Procedure has.
    * For each version:
        * A `uint32` with the index into the Chunks array that corresponds to
          this version of this Procedure.

#### Initial Chunk

* An `uint32`, which is the index to the initial Chunk (Procedure) of a Story.

### Compiled Storyworld Footer

* A 32-bit CRC32 of the payload (using the IEEE polynomial)

### Bits and Pieces

#### Types

##### Built-in Types

Built-in types are represented by a single byte. The value of this byte
determines the type:

* [0] Void
* [1] Boolean
* [2] Int
* [3] Float
* [4] Bounded number
* [5] String

TODO: Do we need to represent Lectures? I don't think so, 

##### User-Defined Types

TODO: User-defined types aren't part of the language yet. Click on the bell to
get notification on Romualdo news. Or whatever.

#### Values

The encoding of a value depends on its type.

##### Boolean Value

Booleans are always represented by one single byte:

* A byte with the value `0` (if `false`), or `1` (if `true`).

##### Int Value

* A byte `2` to indicate it is an `int`.
* An `int64` with the value. The Romualdo spec is (kinda intentionally) vague in
  terms of what's the range of an `int`, but we serialize them as 64-bit values.

##### Float Value

* A byte `3` to indicate it is a `float`.
* Eight bytes containing an IEEE 754 binary64 number (AKA [double precision
  floating
  point](https://en.wikipedia.org/wiki/Double-precision_floating-point_format)).

##### Bounded Number Value

* A byte `4` to indicate it is a `bnum`.
* The value is just like a `float`.

##### String Value

* A byte `5` to indicate it is a `string`.
* An `uint32` with the length of the string in bytes.
* An array of bytes, with the string data encoded in UTF-8. Line breaks are
  represented Unix-style, i.e., by a single LF (`0x0A`) character.

##### Lecture Value

TODO: Do we even need to represent Lectures? There are no Lecture values as
such, as far as I can think.

* A byte `6` to indicate it is a Lecture.
* The value is just like a `string`.

## Debug Info

### Debug Info Header

* An 8-byte "magic number" comprised of the string `RmldDbg` followed by a SUB
  character. These are written to the file in this exact order, i.e., the
  first byte on the file is `R`, the second is `m`, and so on.
* A `uint32` with the version (currently 0).

### Debug Info Payload

#### Number of Chunks

* An `uint32` with the number of Chunks. (This is sort of redundant, because we
  could theoretically get this value from the Compiled Storyworld. Choosing to
  make the Debug Info more self-sufficient and adding this value here too.)

#### Chunks Source Files

This is just like the Chunk Names, but the strings represent the path to the
files from where each Chunk came from. This is the absolute path, from the root
of the Storyworld.

#### Chunks Lines

* For each chunk, we have the mapping from instruction to source code lines. It
  goes like this:
    * A `uint32` with the length of data.
    * This many `uint32`s, each one containing the line number which generated
      that byte of bytecode.

### Debug Info Footer

* A 32-bit CRC32 of the payload (using the IEEE polynomial)

## VM Saved State

### VM Saved State Header

* An 8-byte "magic number" comprised of the string `RmldSav` followed by a SUB
  character (`0x1A`, which in times long gone used to represent a "soft
  end-of-file"). These are written to the file in this exact order, i.e., the
  first byte on the file is `R`, the second is `m`, and so on.
* A `uint32` with the version (currently 0).

### VM Saved State Payload

#### VM State

* An `int32` with the VM state. The value must be:
    * `0` for the "new" state.
    * `1` for "waiting for input".
    * `2` for "end of story".

#### Options

* One string, which looks like this:
    * A `uint32` with the string length.
    * The string data (UTF-8-encoded) with the options.

#### Stack

* One `uint32` with the stack size.
* One Value (as described earlier) for each stack element, from bottom to top.

#### Call frames

* One `uint32` with the number of call frames.
* Each of the call frames, from bottom to top. Each call frame looks like this:
    * An `uint32` with the index of the Chunk corresponding to the call frame's Procedure.
    * An `uint32` with the instruction pointer (IP).
    * An `uint32` with the index into the stack corresponding to the base of the
      stack view used by this call frame.

### VM Saved State Footer

* A 32-bit CRC32 of the payload (using the IEEE polynomial)
