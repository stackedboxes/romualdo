# File Formats

## Compiled Storyworld

All integers are stored in little endian.

### Compiled Storyworld Header

* An 8-byte "magic number" comprised of the string `RmldCSW` followed by a SUB
  character (`0x1A`, which in times long gone used to represent a "soft
  end-of-file"). These are written to the file in this exact order, i.e., the
  first byte on the file is `R`, the second is `m`, and so on.
* A `uint32` with the version (currently 0).

### Compiled Storyworld Payload

#### Constants

* A `uint32` with the number of constants
* Each of the constants, as Values (see below)

#### Chunks

* A `uint32` with the number of Chunks.
* Each of the Chunks, which looks like this:
    * A `uint32` with Chunk size.
    * An array of bytes, with the bytecode. The opcodes and instruction format
      are documented in [Instruction Set](instruction_set.md).

#### Initial Chunk

* An `uint32`, which is the index to the initial Chunk (Procedure) of a Story.

### Compiled Storyworld Footer

* A 32-bit CRC32 of the payload (using the IEEE polynomial)

### Bits and Pieces

#### Values

The encoding of a value depends on its type.

##### Boolean

Booleans are always represented by one single byte:

* A byte with the vaule `0` (if `false`,) `1` (if `true`).

##### Int

* A byte `2` to indicate it is an `int`.
* An `int64` with the value. The Romualdo spec is (kinda intentionally) vague in
  terms of what's the range of an `int`, but we serialize them as 64-bit values.

##### Float

* A byte `3` to indicate it is a `float`.
* Eight bytes containing an IEEE 754 binary64 number (AKA [double precision
  floating
  point](https://en.wikipedia.org/wiki/Double-precision_floating-point_format)).
  We don't mess with endianness in this case: the most significant bit of the
  first byte is the sign, the next ones contain the exponent and the significand
  (mantissa) comes in the later bytes.

##### Bounded Number

* A byte `4` to indicate it is a `bnum`.
* The value is just like a `float`.

##### String

* A byte `5` to indicate it is a `string`.
* An `uint32` with the length of the string in bytes.
* An array of bytes, with the string data encoded in UTF-8. Line breaks are
  represented Unix-style, i.e., by a single LF (`0x0A`) character.

##### Lecture

* A byte `6` to indicate it is a Lecture.
* The value is just like a `string`.

## Debug Info

### Debug Info Header

* An 8-byte "magic number" comprised of the string `RmldDbg` followed by a SUB
  character. These are written to the file in this exact order, i.e., the
  first byte on the file is `R`, the second is `m`, and so on.
* A `uint32` with the version (currently 0).

### Debug Info Payload

* An `uint32` with the number of Chunks. (This is sort of redundant, because we
  could theoretically get this value from the Compile Storyworld. Choosing to
  make the Debug Info more self-sufficient and adding this value here too.)

#### Chunks Names

* One string for each Chunk, each of which looking like this:
    * A `uint32` with the string length.
    * The string data (UTF-8-encoded) with the fully-qualified name of the
      Procedure represented by that Chunk.

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
