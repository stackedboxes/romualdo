# Language Design Notes

First of all, here are the [specs of the previous Romualdo
iteration](https://github.com/lmbarros/romualdo-language/blob/master/doc/maybe-the-ideal-grammar.md).

For this iteration, the main changes are:

* The output from Romualdo to the host program is text only. At least initially.
  And, of course this text could be structured and interpreted by the host in
  various ways -- but from Romualdo's perspective, it's just text.
* Outputting text is much simpler. Basically, within a `passage` things are
  scanned differently: everything is text that will be output. Programming
  constructs are still possible, but keywords must be prefixed by a backslash
  (`\`).

## Ongoing design

### Tree Walk Interpreter

Do I need it?

I want a way to run Romualdo code at compile time (to support compile-time
checkers). But I guess this can even be a second step. Like, compile everything
(including the checker), then running the checker on every Lecture. *If* I find
a way to identify every Lecture in the compiled code.

So, all doable with the bytecode interpreter? An therefore no tree walk
interpreter needed?

Update, September 2023: the Tree-Walk Interpreter is past. I decided to remove
it indeed. Leaving these notes for now, as things are still fluid.

### Comments

Long into the implementation and I realized I haven't written anything about
comments yet! I am currently using `\#` to start comments, but I guess `#` alone
should also work when in code mode, right?

I recall at some point using `\ ` (i.e., backslash-space) to start a comment.
Sounds great in theory (doesn't need to waste any extra character!), but felt
odd. May reconsider it.

A silly advantage of `\ ` to start a comment is that nobody will be allowed to
do the common (but aesthetically wrong!) thing which is not leaving any space
between the comment marker and the comment itself.

Also: do I want block-comments? I guess I can live without them, but I never
thought much about it.

In summary, comments are still a TODO, surprisingly.

### Passages

Tentative example:

```romualdo
passage thePassage(): void

    This is text that is outputted when running the passage. The common
    indentation shared by all lines is removed.

    \if whatever then
        And this is outputted conditionally. BTW, backlashed keywords are always
        valid, but we may not need them always, like shown here. We *know* that
        an expression and then `then` must come after `\if`. Now, maybe a bad
        idea but an `end` with less indentation can be interpreted as an `end`
        instead of text with invalid indentation.
    end
end
```

### Text: Statement or literal? What about functions?

TODO!

In functions, we can use

```romualdo
function f(): void
    say
        And here we are in Text mode!
    end

    say This is text-mode, too! \end
end
```

Almost makes me want to not have distinction between functions and passages at
all! One less indentation level is nice, though.

~~What about this: a text literal has an implicit `say` statement.~~ (Not
true... in an explicit `say` the text literal doesn't have an implicit `say`.)

Can't it be both? Like, a text literal is also a statement that causes the text
to be "said". Clumsy!

TODO: Should I rename "text" to something else? Something that implies "saying"
or "output". Like "discourse" or "uttering". "Speech"? "Lecture" would be nice,
too, because it is slightly derogatory, thus remembering the author to not
overdo!

The thing is: a text literal exists only in text mode. Text isn't really a type.

What about this: we have things called "lectures". A Lecture is automatically
output. The two ways to produce a Lecture are (1) `say` statements, (2)
`passage`s. A Lecture has characteristics of both statements and literals, but
we don't try to fit the concept into the traditional programming languages
lingo.

### Text interpolation

TODO!

```romualdo
Here's some text saying that one plus one equals {1 + 1}. Or: an
expression between curly braces is evaluated, converted to a string and
interpolated into the text.

And this {{domeSomethingForTheSideEffects()}} just some text, right? Or:
arbitrary code between double curly braces is evaluated and the result is
discarded. In this odd example, there would be two spaces between "this" and
"just". Normally, I'd expect double curlies to not be used inline real text.
```

Mnemonic: more curlies don't allow the value to escape!

A good question is: what if someone calls a procedure between curlies, and this
procedure tries to show text?

### Listening to user input

TODO: Think about a longer-term solution.

For the short term, at least, we can do this:

```romualdo
var result: int
result = listen ["alternative 1", "alternative 2", "alternative 3"]
```

That is, `listen` takes an array of strings (the choices offered to the Player)
and returns an integer (the index of the Player choice).

TODO: For the *very first version*, before I support arrays or `int`s:

```romualdo
var result: string
result = listen "What's your favorite color?"
```

### Output filters and checkers

TODO!

If I ever change the output to be something more generic (e.g. a `map`, like in
the previous Romualdo iteration) I can introduce the concept of output filters.
(I actually toyed with this idea before.)

A filter is a procedure (typically a function) that takes the full Lecture
contents (as a string) and transforms it to a `map`.

The default filter would simply take the whole Lecture and put into, say, a
`lecture` field of the output `map`.

Maybe we could even allow to set more than one filter. Like, one to
automatically add the world state to every response, and one to do the actual
conversion. Something like that.

And what about filter-like **checkers**? My Lecture may contain specially
formatted commands that the host may interpret. Like `@@Image: foo` or
something. A checker could check if these commands are correct.

Can we do this a compile-time? Not fully, because of interpolated contents in
Lectures. But to some extent maybe? Let's say we replace all interpolated bits
with a hardcoded string (or just leave the curly-expressions as they are) before
running the compile-time checker. I think a well-designed checker along with a
well-designed Lecture format (and not being too creative with interpolations)
could work well-enough to catch pretty much all relevant errors at compile-time.

### Unit tests

I'll probably want to add some simple unit testing facility, inspired by D
(though maybe even simpler). Something like this:

```romualdo
unittest
    std.assert(doubleIt(3) == 6)
    std.assert(doubleIt(7) == 14)
end
```

That is, `unittest` blocks at global level are usually ignored (though they are
parsed and checked). But when running with the right `romualdo` command, each of
these blocks is executed (and `main` isn't).

`unittest` blocks should probably reset all globals before running themselves.
Or just start a brand new interpreter for each block and run them in parallel.

## Passages x Functions

In principle, both should be allowed to do the same things. It's just that the
syntax accepted by each one is different, favoring either text or code.

### Code in Passages

```romualdo
passage p(): void
    Alright, so this is some lecturing we are making
    \if wantMoreLecturing == "a lot" then
        This is more, much more, much more, much more,
        much more, much more, much more, much more,
        much more, much more, much more, much more,
        much more, much more, much more, lecturing
        for ya!
    \else      \# backslash actually optional here, because of dedenting
        Alright, I am stopping here!
    \end       \# ditto
end

This is equivalent to:

```romualdo
function p(): void
    say
        Alright, so this is some lecturing we are making
    end
    if wantMoreLecturing == "a lot" then
        This is more, much more, much more, much more,
        much more, much more, much more, much more,
        much more, much more, much more, much more,
        much more, much more, much more, lecturing
        for ya!
    else
        say
            Alright, I am stopping here!
        end
    end
end
```

Passages simply are Lecture mode by default and `say`/`end` is implicit.

### Curlies and Passages

What if we run some code in curlies that eventually runs `say`? This should be
forbidden, right? This will just lead to confusing, wrong, unintentionally
interleaving output from different Passages.

So, how can we avoid it?

At compile time we can check if there is *any* chance of a `say` being executed.
Not sure this can be foolproof (at least not without forbidding a lot of code
that is actually harmless -- AKA false positives).

So, if we hit a `say` while on curlies, skip the `say` ("no runtime errors") and
log it or whatever. But what else? Is this enough?

Some sort of coloring (procedures are "colored" as either talky or silent;
silents cannot call talkies) would technically work. But that's basically what I
could achieve with static analysis (the compile-time checks I mentioned above).
It's likely to end up making the vast majority of procedures talkies and thus
forbidding their usage on curlies. *But...* is this bad, really? Effectively,
what I am doing is saying that only procedures that are very clearly silents can
be called from curlies. Curlies are meant to do relatively simple things. So,
maybe that's my solution after all.

### Arrays and maps

First big challenge here is: how to avoid runtime errors? Out-of-bounds indices,
nonexisting keys, bad types. I want means to check for these conditions before
trying them, of course. But I need a proper solution for *when* they happen! The
obvious solution is to always have a default return value. But with what syntax?

I mean, returning the default value for the type would work (and maybe is a
decent default/fallback), but would be nice to offer a way to the user to
provide the in-case-of-error value.

Let's try:

```romualdo
var i = myArray[3]!171
var s = myMap["key"]!"default"
```

Anyway, I'd say accessing a wrong index without providing a default *is* a sort
of "soft error", and I would like to warn/log it in test runs or rehearsal.

Not bad. What about writing? Writing to a nonexisting map key just creates the
new entry. Fine. No so easy with arrays! I don't think I can do any better than
a no-op and soft error.

TODO: Can I make chains have a single default value? Like this:

```romualdo
var i = myArray[3][5][1]!171              \# Uses default if any of the accesses fail
var s = myMap["key1"]["key2"]!"default"   \# Ditto
```

### Modules (Packages?)

Scratch all that comes below after the horizontal line. It's probably a better
idea to require something like:

```romualdo
import "foo/bar" as myBar
import "foo/bar"                     \# imports as bar by default
import "/foo/bar"                    \# absolute path (from Storyworld root)
import "../bar"                      \# importing from parent module
import "/module with spaces" as mws  \# bad, but not right; `as` alias needed in this case
```

TODO: Do I even need the concept of fully-qualified names, then? Perhaps only
for error messages and the like? Well, the binaries and interpreter will need a
FQN to avoid ambiguities.

TODO: FQNs could use a slash, too. `/main` is the entry point.
`/passages/chapter_1/happy_transition` could be a Procedure called
`happy_transition` at the `/passages/chapter_1/` Package.

TODO: Should I forbid modules with spaces in their names? Probably! In which
case, the import syntax wouldn't need to get a quoted string.

TODO: Module or package? Package seems to be more like what Go, Python and Java
would use for what I have here. (The Lua book says: "The `package` library
provides basic facilities for loading modules in Lua.")

----

A Storyworld is made of **Modules**. Each Module resides in a directory, and the
**Main Module** is at the root directory of your Storyworld. All other Modules
are in fact **Submodules** of the main one. The name of the Main Module is
`main`. The name of a Submodule is the name of the subdirectory where it
resides, but no Module other than the Main Module other module can be called
`main`, regardless of the nesting depth.

Submodules can nest deep as you want, that is to say, you can have a Submodule
of a Submodule, of a Submodule, etc. You can use periods to join Module names
and thus build complete Module paths, like
`main.submodule1.submodule2.submodule3`.

**TODO:** Above: define "module paths". Or call it qualified name or something
like that. FWIW, I use the term "Fully Qualified Name" below.

If the first component of a Module path is `main`, we know that it is an
absolute path. Otherwise, it is interpreted as relative from the current Module.

**TODO:** Support relative "up-paths", like the `..` in file systems? With which
syntax?

Each Module can be implemented in a single or in multiple files. Splitting a
Module into multiple files is just a matter of convenience: it makes no
difference from the perspective of the language.

### Meta blocks

For top-level procedures, needed because they act as static vars. So, I need a
syntax for that; and one that works with Passages, too.

```romualdo
function f(): void
    meta
        var v: bool = true
    end
end

passage p(): void
    \meta
        var execCount: int = 0
    \end
    {{ execCount += 1 }}
    And here finally there is some Lecture.
end
```

So, the meta isn't versioned, only the Procedure is.

At global-level, I used to say meta would be the way to set the Storyworld
version. But... do I need this? Why?

I really need the version only in exported bytecode (and even then, only for
debugging and informational purposes, AFAICT). Actually, better to generate a
version only when using some special compiler flag or command. Otherwise, for
day-to-day work, generate binaries with, say, negative versions, indicating they
are WIP.

TODO: How to read a `meta`? Maybe `package.Func.metaName`. Int his case, need to
allow the "third segment" in assignments and reads: `qualifiedIdentifier` alone
won't do.

## How to avoid infinite loops?

A big, big TODO!

Each amd every Procedure call can potentially recurse infinitely and cause a
crash. It's easy to add a size limit to the call stack and thus replace the
crash with a runtime error... but I want to avoid runtime errors!

How can I?

Maybe every procedure needs a default return value?

And configurable max stack size? (Configurable in stack frames.)

Interesting corner case: what if I implement tail call elimination? For tail
calls we'd never reach the stack limit, but I'd still like to avoid the infinite
loop! Maybe tail call elimination is bad for Romualdo?

On the other hand, infinite loops in iterative code are equally bad, and have
nothing to do with recursion or tail calls. Can I avoid infinite recursion by
allowing only "very well-behaved" loops? Like iterating over the elements of an
immutable collection? Sounds too restrictive!

## Ink-like Variable text?

One thing I have a kind of [Ink](https://www.inklestudios.com/ink/)-envy is
their set [variable
text](https://github.com/inkle/ink/blob/master/Documentation/WritingWithInk.md#8-variable-text)
features. Like this, taken directly from their docs:

```ink
The radio hissed into life. {"Three!"|"Two!"|"One!"|There was the white noise racket of an explosion.|But it was just static.}
```

This kind of thing is definitely not my focus with Romualdo (again, roughly
speaking, Romualdo leans more towards programming than text generation), but I
think these are cool features.

How could I support this? Imagining something like this:

```romualdo
say
    The radio hissed into life. { say.sequence("\"Three!\"", "\"Two!\"", "\"One!\"", "There was the white noise racket of an explosion.", "But it was just static.") }
end
```

Two problems. First, ugly quote escaping. Second, `say.sequence()` needs to
somehow store state *per call site.*

What about an implicit *call site* argument available to every Procedure? Or
explicitly declared by Procedures that want it, but implicitly passed. Or
something like this:

```romualdo
function foo(): void
    var percallsite i = 0
    \# ...
end
```

Could such feature be useful for other useful stuff beyond variable text?

## "Multithreading"

This is food for thought for a distant future. The Romualdo equivalent to
multithreading would be a story with several parallel, well, threads, going on.

Should be technically doable, but I am not so sure it is useful enough. Maybe
there are simpler ways to achieve the same goals.

## Principles

Rough and kinda conflicting, but these are some principles I am trying to
follow. These are more about implementation than design.

* **Maintainability over performance.** At least within reasonable limits. I am
  creating this because I want to use it, not because I want to maintain it.
* **User friendliness over best practices.** I don't mind having two almost
  identical functions if they can provide better error messages (compared with
  merging them into a single function). "User" in this case means "me".
* **No runtime errors.** Stories don't crash, and stories is what we are trying
  to make.
