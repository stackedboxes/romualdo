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

### Versioning

(This is largely not implemented, but here's how I am planning to do this.)

Imagine you released a Storyworld that takes a while to play through. Someone
plays it for hours and then saves their current progress. At this point you
release an update to the Storyworld and your player, mouth watering for the
shiny new features, immediately updates -- only to find that the saved progress
is not compatible with the new release and he needs to start from the beginning.
How frustrating!

Versioning exists to avoid this kind of scenario. It enables compatibility
between old saved states and new Storyworld releases.

Versioning is a feature of the Romualdo tool, not of the Romualdo language. In
other words, it's all about how you invoke the `romualdo` tool -- you don't need
to change anything to the code you write. (Though, I guess the way you write
your code may affect how the user "perceives" versioning. TODO: Explore this,
add examples! Like, a long main procedure with a hardcoded ending versus a short
one calling and `end()` procedure; the first case will not see an updated
ending; the second will.)

#### Technical overview

Under the hood, what we do is conceptually simple. We keep the code for all
released versions of all procedures in the Storyworld. So, if a saved state
refers to an old version, it can still run it because the code is there! But
whenever a procedure is called, it's the latest version that gets executed.

I think this is not too different from what Erlang does to support hot updates
(i.e., to deploy a new version while the old version is running). In the case of
Erlang, they can dispose the old executable code once nobody is running it on
the VM anymore, while Romualdo needs to keep the old code around forever,
because old saved states are eternal. (I never used Erlang seriously, and the
one time I used it was some 20 years ago, so my recollection might be a bit
off!)

#### Open points

**TODO:** How to deal nicely with changing signatures? Probably just forbid for
now. Old code may need to call new procedures, and for that to be possible they
need to have the same signature. I guess it could be OK to add arguments with
default values (if I ever support that). And: completely changing the meaning of
arguments in a new version of a procedure will break the semantic of saved
states!

#### Tutorial-like description

[This is tutorial-like but I am adding some notes about the internals in square
brackets.]

The first time you

```sh
romualdo build PATH
```

the compiler generates a (say) `red_hoodie.csw` file in which everything is
internally marked as being unreleased. In other words, all this compiled code is
considered a development version, not something players should put their hands
on.

[Internally, things are marked as release `-1`; real releases index an array of
releases, so a negative value means that these things aren't really part of a
real release.]

[Also, internally, every procedure is hashed, and the hash is saved to the
`.csw` file. This allows to check for compatibility when loading a saved state.
As we'll soon see, this also allows us to check if we need to create a new
version of something when creating new releases.]

So you keep working on your Storyworld, making changes, `build`ing it, and
testing it. Everything will remain internally marked as unreleased. When you are
happy and ready to release, you

```sh
romualdo release PATH VERSION
```

which still generates a `red_hoodie.csw` file, but this one will be your first
release: everything in it is marked as being part of the version you passed as
argument. And it will be a proper release that you can give to your players.

The `VERSION` argument can be any string without spaces.

[At this point we create the first release internally: `0`. We create the
association between this internal release number `0` and the passed `VERSION`,
and mark everything as being on version `0`.]

[A release basically marks things in the `.csw` as immutable. Subsequent
releases will not overwrite any binary code that has a non-negative release
number. These need to be kept forever because some user may have a saved state
referring to it.]

Now you should **commit your `red_hoodie.csw` to version control**: you'll need
it to create future releases! This is actually the perfect moment to commit all
your source code to your version control system and tag this repo state with the
same `VERSION` you passed to `romualdo release`. Why? Because otherwise you'll
not be able to check the source code that corresponds to this release. You may
need to look into the code for an old release only rarely, but when you do need
it, you need it! (And it also adds a layer of protection against Romualdo bugs.
If the `romualdo` tool corrupts you `.csw` file, you can roll-back to a known
good state.)

Time passes, and you decide to change or add something. You update the code and
run

```sh
romualdo build PATH
```

The `romualdo` tool will produce a new `red_hoodie.csw` file that you can use
normally for testing, but which you shouldn't distribute to your players,
because all the new stuff added to it is internally marked as unreleased.

[Internally, things that have been released in release `0` remain unchanged and
marked as release `0`. But changed stuff will get a new copy internally, marked
as version `-1`. Likewise, new stuff is added as version `-1`.]

Next time you `romualdo release`, you'll get an updated `red_hoodie.csw` with
nothing marked as unreleased. That will be a new release (with the version you
passed in) that is ready to be shipped to players.

The `romualdo` tool will bark if there are no changes from the last release. It
will also bark if the previous `red_hoodie.csw` is not available (I said to
version control it!). It can't really know if the `red_hoodie.csw` available is
the right one; it is your responsibility to guarantee that it contains
everything from the last released version. (It can contain unreleased stuff in
addition to that -- that's not a problem.)

And there you have it, your second release! Users can update to it, and their
old save states will keep working normally. (Fine print: limitations may apply!)

[As always, after a `romualdo release` everything on your `.csw` file will be
associated with a released version. No `-1` versions there!]

What about **globals and versioning**?

Well, because of versioning, the Romualdo tool will not allow you to make
certain changes to globals between releases:

1. Cannot remove global variables.
2. Cannot rename global variables.
3. Cannot change the type of global variables.

That's because Procedures in old releases (that may be still used because of
saved states) may depend on the old definitions.

Note that changing the initialization expression of a global variable is fine,
because it is used only when starting a new Story, or initializing a global just
added to a new release.

**TODO:** This is bad for external libraries. The whole control of globals is
made at the Storyworld level, so an external library cannot ever remove globals
without this being a breaking change. Reusable libraries would need to release
new major versions whenever they want to remove old globals that aren't used
anymore. And old Storyworlds depending on that library would not be able to
update without breaking compatibility with old saved states. Not great, but I
think I can live with that for now.

So, let's say the first release of the Storyworld had this on a certain package:

```romualdo
globals
    EndGame: bool = false
    artifactsCount: int = 0
end
```

Then a second release does this:

```romualdo
globals
    EndGame: bool = false              \# Fine, same as the first release
    artifactsCount: int = 1            \# Fine, just initialization changed
    favoriteColor: string = "blue"     \# Fine, a brand new global variable
end
```

And we try this on a third release:

```romualdo
globals
    EndGame: string = "sure!"          \# Error! Changed the type
                                       \# Error! Removed some existing globals
end
```

[Internally, along with each global we store it's hash, which is based on its
FQN and type. When releasing, every global hash in the CSW must still be present
on the source.]

#### Compatibility between saved states and compiled Storyworlds

Finally, let's talk about the compatibility between a saved state and a given
compiled Storyworld.

First off the entries in the call stack (which ends up replicated in a saved
state) must include (directly or indirectly) the hash of the version of each
running procedure. This is what allow us to use an old version if that was what
the user was running.

When loading a saved state, the loader checks for compatibility between the
saved state and the compiled Storyworld being used. The algorithm for that is
the following.

* For each procedure `p` in the call stack of the saved state:
    * If `p.hash` does not appear in the compiled Storyworld, they are
      incompatible.
* For each global variable `g` in the saved state:
    * If `g.hash` does not appear in the compiled Storyworld, they are
      incompatible.
* If we reach this point, they are compatible!

Since this is all based on hashes of the actual code of the procedures adn
globals, this algorithm works both for:

1. Released cases. For example, I send a saved state to a friend, but the friend
   is still using an older version of the Storyworld. They need to update the
   Storyworld, otherwise loading will (rightly) fail.
2. Unreleased cases. This is useful for testing. If I am testing an unreleased
   version of the Storyworld I can save the state and load it multiple times to
   test different things, even with changes to the Storyworld -- as long as I
   don't change any of the unreleased procedures currently on the call stack.

#### Hashing Procedures and global variables

The hash of a procedure is computed like this:

* Initialize an MD5 hash computation.
* For each token of the procedure, from `function` or `passage` to `end`
  (inclusive at both ends):
    * Add to the hash computation the token lexeme (that is exactly what we have
      in `frontend.Token.Lexeme`, which may already include some cleaning,
      especially for Lectures).
    * Add the a zero byte to the hash computation. This a single byte with all
      bits set to zero, not a string with an ASCII "0" character.

Note that by taking into account only the tokens, we allow changes to formatting
and comments (which do not affect the generated code).

MD5 should be a good choice here. It should be fast enough and security is not a
concern here.

For global variables, it's similar in concept, but with one tricky detail. We
want to take into account the variable name and type and ignore the initializer
-- but the type can be syntactically omitted if there is an initializer, and in
this case we need to "manually" include the inferred type, because we really
need to have the type as part of the hash.

Alternative: compute on the parser would be significantly more complex (with
tokens getting asked for everywhere throughout the parser)

Alternative: compute on the AST. Main issue here is that we'd lose the ability
to change the grammar, as that would change the hash. Worth checking if the AST
gives us enough info (and, more than that, in a convenient format) to compute
what would be effectively the same thing as the scanner-based approach.

**TODO:** We are ignoring the package imports, which is fine. But then, suppose
someone changes a package import to rename the imported name. They need to
change the code accordingly. But this also doesn't change the generated code,
because it's all resolving to the FQN under the hood. So, we could consider this
kind of change as non-breaking. This is not a terrible problem, really, but
could be a pro for an approach based on hashing the AST (at an early stage,
before transformations or lowerings.)

Implementation-wise, for Procedures we'll let the scanner do the actual hash
computation, with the parser telling the scanner when to reset the computation.
The parser is also the one asking the scanner for the hash, whenever it needs
it. For global variables it's probably much, much simpler to do at the AST level
(at scanner level we don't know about inferred types). **TODO:** I hate the
difference in handling of Procedures and globals.

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
