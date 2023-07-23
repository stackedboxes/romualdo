# The Romualdo Virtual Machine Instruction Set

Not yet assigning a definitive value (or, er, "byte code") to each instruction,
but let's at least document what we can do.

## Assorted Topics

### Unbounded and Bounded Numbers

TODO: This section is theoretical, this is not implemented yet.

Romualdo (both the language and the VM) has three types of numeric values:
`int`s, `float`s and `bnum`s. Whenever we mention "unbounded numbers" in this
document, we are talking about `int`s and `float`s. This is in contrast with
`bnum`s which are "bounded numbers".

### Operations Between Different Types

TODO: This section is theoretical, this is not implemented yet.

Essentially, the behavior of the VM matches the behavior of the language. In
general, operations between different types are not supported and values of
different types are considered different.

The only exception is when one operand is an `int` and the other is a `float`:
in this case, the `int` one is converted to a `float` and then the operation is
performed as if both operands were `float`s.

There is no automatic type conversion like this for any other type, not even for
bounded numbers.

All arithmetic operations between `float`s result in a `float`.

All arithmetic operations between `bnum`s result in a `bnum`.

Most arithmetic operations between `int`s result in `int`s. The exceptions are
`DIVIDE` and `POWER`, which always yield `float` results.

### Immediate operands

TODO: This section is theoretical, this is not implemented yet. And in fact I
must rethink if we'll have any byte-sized immediate operands. Going full 32-bit
may simplify things a good deal.

Each instruction that has immediate operands interpret them in one of the few
possible ways described below. The description of each instruction tells which
of these interpretations it uses.

* **Unsigned byte.** The operand is a single byte, interpreted as an unsigned
  integer.
* **Signed byte.** The operand is a single byte, interpreted as a signed integer
  encoded in two's complement.
* **Unsigned 32-bit integer.** The operand is a 32-bit unsigned integer, stored
  in little-endian format.
* **Signed 32-bit integer.** The operand is a 32-bit signed integer, stored in
  little-endian byte order, encoded in two's complement.

### Calling convention

TODO: This section is theoretical, this is not implemented yet.

When a Procedure (the caller) calls another Procedure (the callee), what happens
is the following.

1. The caller pushes into the stack the Procedure object representing the callee.
2. The caller pushes into the stack any arguments required by the callee. The
   arguments are pushed in the same order they appear in the callee Procedure
   declaration. (In other words, push the first argument first, then the second
   one, and so on.) If the callee doesn't take any arguments, this step is a
   no-op.
3. The caller executes the `CALL` instruction. This passes the control to the
   callee.
4. The callee does it's stuff. The VM will set the callee's stack such that
   index 0 will contain the Procedure object representing the callee, index 1
   will contain the first argument, index 2 the second argument and so on.
5. If the callee returns a non-void value, it pushes the return value into the
   stack and calls `RETURN_VALUE`. If the called returns void, it calls
   `RETURN_VOID` (without pushing anything).
6. In either case, the execution of the `RETURN_*` opcode will pop all its
   locals and arguments (but will keep the return value on the top of the stack,
   if there is a return value).
7. The control passes back to the caller.

This is not something enforced by the virtual machine (VM) itself but rather, as
the name implies, a convention. I'd say that it's generally a good idea to
follow it, though. Don't try to outsmart the VM.

As always, by Procedure we mean either a Function or a Passage.

## The Instructions

Instructions are listed in alphabetical order.

The fields "Pops" and "Pushes" describe the effects as perceived by the user,
not necessarily how the implementation works. For example, if you see some
instruction that pops a value and then pushes the same value back to the stack,
the implementation is free to leave the stack untouched.

### `CONSTANT`

**Purpose:** Loads a constant with index in the [0, 255] interval.  
**Immediate Operands:** One byte *A*, interpreted as an index into the constant
pool.  
**Pops:** Nothing.  
**Pushes:** One value, the value of constant taken at the index *A* of the
constant pool.

### `EQUAL`

**Purpose:** Checks if two values are equal.  
**Immediate Operands:** None.  
**Pops:** Two values, *B* and *A*.  
**Pushes:** One Boolean value telling if *A* = *B*.

### `FALSE`

**Purpose:** Loads a `false` value.  
**Immediate Operands:** None.  
**Pops:** Nothing.  
**Pushes:** One Boolean value: `false`.

### `JUMP`

**Purpose:** Jumps to a different location unconditionally.  
**Immediate Operands:** One signed 32-bit integer, *A*, interpreted as the
offset to jump.  
**Pops:** Nothing.  
**Pushes:** Nothing.  
**Other Effects:** Sets the instruction pointer to a value equals to the
instruction address, plus *A*.

Notice that *A* is signed, so we can jump forward or backward. If *A* is zero,
we'll have an infinite loop.

By the way, the jump offset is designed to be an offset from the instruction
address itself because I find it easier to reason about when looking at
disassembled code. (I mean, compared with the arguably more common alternative
of using the address of the next instruction as the jump base).

### `JUMP_IF_FALSE`

**Purpose:** Jumps to a different location maybe.  
**Immediate Operands:** One signed 32-bit integer, *A*, interpreted as the
offset to jump.  
**Pops:** One Boolean value *A*.  
**Pushes:** Nothing.  
**Other Effects:** If *A* is a Boolean value and is false, sets the instruction
pointer to a value equals to the instruction address, plus *A*.

Notice that *A* is signed, so we can jump forward or backward. *A* can
technically be zero, but there's no obvious high-level semantic for that: the
instruction will keep executing itself as long the stack top contains `false`
values, then will proceed to the next instruction.

### `LISTEN`

TODO: This string-based interface is temporary, until we support richer types.

**Purpose:** Pauses the execution and waits for user input.  
**Immediate Operands:** None.  
**Pops:** One value, a string with the options to show to the user.  
**Pushes:** One value, a string with the Player choice.

A `LISTEN` instruction pauses the execution and returns control to the driver
program. At this point, the VM should have popped the options string from the
stack. When the driver program resumes the VM execution, the Player choice
string will be pushed, so that the next instruction will have access to it
already.

### `NOP`

**Purpose:** Does nothing.  
**Immediate Operands:** None.  
**Pops:** Nothing.  
**Pushes:** Nothing.

I can't really see any purpose for a no-op instruction in the Romualdo VM, but I
*really* wanted to have it. That's probably because of the tender memories I
have of `NOP` in the x86 architecture. Whatever.

### `NOT_EQUAL`

**Purpose:** Checks if two values are different.  
**Immediate Operands:** None.  
**Pops:** Two values, *B* and *A*.  
**Pushes:** One Boolean value, telling if *A* ≠ *B*.

Equivalent to `EQUAL` followed by `NOT`, but in a single, efficient instruction.

### `POP`

**Purpose:** Pops the value on the top of the stack.  
**Immediate Operands:** None.  
**Pops:** One value.  
**Pushes:** Nothing.

### `SAY`

**Purpose:** Sends the contents of a Lecture to the Driver Program.  
**Immediate Operands:** None.  
**Pops:** One value, the Lecture to be said.  
**Pushes:** Nothing.

### `TO_LECTURE`

**Purpose:** Converts a value to a Lecture.  
**Immediate Operands:** None.  
**Pops:** One value, *A*, the value to convert to a Lecture.  
**Pushes:** One value: *A* converted to a Lecture.

This is a no-op if the value at the top of the stack is already a Lecture.
Otherwise, this is essentially the same as `TO_STRING`, but leaving a Lecture on
the stack (instead of a string).

### `TO_STRING`

**Purpose:** Converts a value to a string.  
**Immediate Operands:** None.  
**Pops:** One value, *A*, the value to convert to a string.  
**Pushes:** One value: the string representation of *A*.

### `TRUE`

**Purpose:** Loads a `true` value.  
**Immediate Operands:** None.  
**Pops:** Nothing.  
**Pushes:** One Boolean value: `true`.
