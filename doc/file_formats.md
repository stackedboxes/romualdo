# File Formats

## Compiled Storyworld

All integers are stored in little endian.

## Header

* An 8-byte "magic number" comprised of the string `RmldCSW` followed by a SUB
  character (`0x1A`, which in times long gone used to represent a "soft
  end-of-file"). These are written to the file in this exact order, i.e., the
  first byte on the file is `R`, the second is `m`, and so on.
* A `uint32` with the version (currently 0).

## Payload

### Constants

* A `uint32` with the number of constants
* Each of the constants, as Values (see below)

### Chunks

* A `uint32` with the number of Chunks.
* Each of the Chunks, which looks like this:
    * A `uint32` with Chunk size.
    * An array of bytes, with the bytecode. The opcodes and instruction format
      are documented in [Instruction Set](instruction_set.md).

### First Chunk

* An `uint32`, which is the index to the first chunk (Procedure) of a Story.

## Footer

* A 32-bit CRC32 of the payload (using the IEEE polynomial)

## Bits and Pieces

### Values

The encoding of a value depends on its type.

#### Boolean

Booleans are always represented by one single byte:

* A byte with the vaule `0` (if `false`,) `1` (if `true`).

#### Int

* A byte `2` to indicate it is an `int`.
* An `int64` with the value. The Romualdo spec is (kinda intentionally) vague in
  terms of what's the range of an `int`, but we serialize them as 64-bit values.

#### Float

* A byte `3` to indicate it is a `float`.
* Eight bytes containing an IEEE 754 binary64 number (AKA [double precision
  floating
  point](https://en.wikipedia.org/wiki/Double-precision_floating-point_format)).
  We don't mess with endianness in this case: the most significant bit of the
  first byte is the sign, the next ones contain the exponent and the significand
  (mantissa) comes in the later bytes.

#### Bounded Number

* A byte `4` to indicate it is a `bnum`.
* The value is just like a `float`.

#### String

* A byte `5` to indicate it is a `string`.
* An `uint32` with the length of the string in bytes.
* An array of bytes, with the string data encoded in UTF-8. Line breaks are
  represented Unix-style, i.e., by a single LF (`0x0A`) character.

#### Lecture

* A byte `6` to indicate it is a Lecture.
* The value is just like a `string`.
