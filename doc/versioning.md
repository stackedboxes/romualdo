# Versioning in Romualdo

***Warning:** This is all tentative, incomplete, and work-in-progress!*

Imagine you released a Storyworld that takes a while to play through. Someone
plays it for hours and then saves their current progress. At this point you
release an update to the Storyworld and your player, mouth watering for the
shiny new features, immediately updates -- only to find that the saved progress
is not compatible with the new release and he needs to start from the beginning.
How frustrating!

Versioning exists to avoid this kind of scenario. It enables compatibility
between old saved states and new Storyworld releases -- at least to a certain
extent.

Versioning is a feature of the Romualdo tool, not of the Romualdo language. In
other words, it's all about how you invoke the `romualdo` tool -- you don't need
to change anything to the code you write. (Though, the way you write your code
may affect how the user "perceives" versioning. We discuss a bit about this
[further down](#dark-corners).)

## Technical overview

Under the hood, what we do is conceptually simple. We keep the code for all
released versions of all procedures in the Storyworld. So, if a saved state
refers to an old version, it can still run it because the code is there! But
whenever a procedure is called, it's the latest version that gets executed.

I'd guess this is not too different from what Erlang does to support hot updates
(i.e., to deploy a new version while the old version is running). In the case of
Erlang, they can dispose the old executable code once nobody is running it on
the VM anymore, while Romualdo needs to keep the old code around forever,
because old saved states are eternal. (I never used Erlang seriously, and the
one time I used it was some 20 years ago, so my recollection might be a bit
off!)

## Tutorial-like description

*[This is mostly tutorial-like but I am adding some notes about the internals
between square brackets. Eventually I want to write separate docs for the
internals and end users, but this mishmash will do for now.]*

The first time you

```sh
romualdo build PATH
```

the compiler generates a (say) `red_hoodie.ras` file in which everything is
internally marked as being unreleased. In other words, all the compiled code is
considered a development version, not something players (your "end users")
should put their hands on.

*[Everything in the Compiled Storyworld is marked as unreleased. Also,
internally, every procedure and global variable is hashed, and the hash is saved
to the `.ras` file. This allows to check for compatibility when loading a saved
state. As we'll soon see, this also allows us to check if we need to create a
new version of something when creating new releases.]*

So you keep working on your Storyworld, making changes, `build`ing it, and
testing it. Everything will remain internally marked as unreleased. When you are
happy and ready to release, you

```sh
romualdo release PATH TAG
```

which still generates a `red_hoodie.ras` file, but this one will be your first
release: everything in it is marked as being part of the version you passed as
argument. And it will be a proper release that you can give to your players.

The `TAG` argument can be any non-empty string without spaces.

*[At this point we create the first release internally: `0`. We create the
association between this internal release number `0` and the passed `TAG`,
and mark everything as being on version `0`.]*

*[A release basically marks things in the `.ras` as immutable. Subsequent
releases will not overwrite any binary code that has a non-negative release
number. These need to be kept forever because some user may have a saved state
referring to it.]*

Speaking of which, now you should **commit your `red_hoodie.ras` to version
control**: you'll need it to create future releases! This is actually the
perfect moment to commit all your source code to your version control system and
tag this repo state with the same `TAG` you passed to `romualdo release`. Why?
Because otherwise you'll not be able to check the source code that corresponds
to this release. You may need to look into the code for an old release only
rarely, but when you do need it, you need it! (And it also adds a layer of
protection against Romualdo bugs. If the `romualdo` tool corrupts you `.ras`
file, you can roll-back to a known good state.)

Time passes, and you decide to change or add something. You update the code and
run

```sh
romualdo build PATH
```

The `romualdo` tool will produce a new `red_hoodie.ras` file that you can use
normally for testing, but which you shouldn't distribute to your players,
because all the new stuff added to it is internally marked as unreleased.

*[Internally, things that have been released in release `0` remain unchanged and
marked as released. But changed stuff will get a new copy internally, marked as
unreleased. Likewise, new stuff is added as unreleased.]*

Next time you `romualdo release`, you'll get an updated `red_hoodie.ras` with
nothing marked as unreleased. That will be a new release (with the version you
passed in) that is ready to be shipped to players.

The `romualdo` tool will bark if there are no changes from the last release. It
will also bark if the previous `red_hoodie.ras` is not available (I said to
version control it!). It can't really know if the `red_hoodie.ras` available is
the right one; it is your responsibility to guarantee that it contains
everything from the last released version. (It can contain unreleased stuff in
addition to that -- that's not a problem.)

And there you have it, your second release! Users can update to it, and their
old save states will keep working normally. (Fine print: limitations may apply!)

*[As always, after a `romualdo release` everything on your `.ras` file will be
marked as released.]*

One final thing I'd like to note here (and deserves more detailed docs --
they'll come eventually!) is that if you change an existing Procedure and build
or release your Storyworld, three different can happen:

* *A new version is generated.* This is the normal case: you change the
  Procedure, it gets a new version.
* *Nothing.* Like, the same version is used. Yes, there are change changes that
  will not cause a new version to be created. For example, it is fine to change
  spacing and indentation, even in some cases to add or remove some redundant
  parenthesis.
* *You get an error.* There are changes you simply cannot make to a Procedure,
  and the compiler will bark to let you know if you try any of these. The one
  forbidden change that comes to mind is changing the Procedure's argument list
  and/or return value.

What about **versioning of global variables**?

Well, because of versioning, the Romualdo tool will not allow you to make
certain changes to globals between releases:

1. Cannot remove global variables.
2. Cannot change the type of global variables.
3. Cannot rename global variables. (Well, renaming is interpreted as adding a
   new global, which is fine, and removing a global, which is not.)

That's because Procedures in old releases (that may be still used because of
saved states) may depend on the old definitions.

Note that changing the initialization expression of a global variable is fine,
because it is used only when starting a new Story, or initializing a global just
added to a new release.

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

And on a third release we try this:

```romualdo
globals
    EndGame: string = "sure!"          \# Error! Changed the type
                                       \# Error! Removed some existing globals
end
```

*[Internally, along with each global we store it's hash, which is based on its
fully-qualified name and type. When releasing, every global hash in the Compiled
Storyworld must still be present on the source.]*

## Compatibility between saved states and compiled Storyworlds

Let's talk about the compatibility between a saved state and a given compiled
Storyworld.

First off, each entry in the call stack (which ends up in the saved state) must
include the hash of the exact version of the Procedure that was running. This
hash allows us to notice that the user was running an old version and, since the
`.ras` file keeps the bytecode for old versions of all released procedures, we
can just use it.

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

Since this is all based on hashes of the actual code of the procedures and
globals, this algorithm works both for:

1. Released cases. For example, I send a saved state to a friend, but the friend
   is still using an older version of the Storyworld. They need to update the
   Storyworld, otherwise loading will (rightly) fail.
2. Unreleased cases. This is useful for testing. If I am testing an unreleased
   version of the Storyworld I can save the state and load it multiple times to
   test different things, even with changes to the Storyworld -- as long as I
   don't change any of the unreleased procedures currently on the call stack.

## Hashing Procedures and Global Variables

We use SHA-256 as the underlying hashing algorithm. This is probably an
overkill: MD5 should probably have been enough for this use case, but I was
paranoid about possible collisions, so I went completely overboard and doubled
the size of the hash.

### Hashing Procedures

To give you a first approximation of how we compute a Procedure hash, take this
algorithm:

* Initialize a SHA-256 hash computation.
* For each token of the procedure, from `function` or `passage` to `end`
  (inclusive at both ends):
    * Add the token to the hash computation.
    * Add a zero byte to the hash computation (this is a single byte with all
      bits set to zero, not a string with an ASCII "0"
      character!).[^zero-after-token]
* Complete the SHA-256 computation. The result is the Procedure hash.

[^zero-after-token]: The zero byte after each token is there just in case, to
  disambiguate between two consecutive tokens that could be interpreted as a
  single different token. For example, this makes sure tokens `else` and `if`
  are hashed to a different value than the single token `elseif`. (Though this
  is example is only theoretical; for reasons mentioned below, the `elseif`
  token never shows up when computing code hashes.)

In practice, the way we really compute hashes departs a bit from this idealized
algorithm. There are two reasons for this. First, simply looking at the tokens
as they appear in the source code would not work in certain cases. And second,
for simplicity, the code hashing implementation actually works on the AST
(abstract syntax tree) and not on the raw token stream.

Here's the list of known differences between the idealized algorithm I showed
first and the actual algorithm implemented:

* The token lexeme we use is what we have in `frontend.Token.Lexeme`. This is
  not exactly the same thing we see in the source code. Notably, Lectures will
  have indentation removed.
* All symbols appearing in the code are replaced with their fully-qualified
  names. This is meant to avoid a corner-case-ish problem: imagine you import
  package as `import /util/random as r`. Your Procedure code will refer to
  things like `r.drawInt()`. But then you change your import to `import
  /super_util/random as r`. This is a breaking change, but your Procedure code
  didn't change. Using the fully-qualified name will detect the breaking change.
* All `elseif`s are transformed into full-fledged `else if`s (including the
  added `end` to close the new standalone `if`). This is so because in the AST
  there are no `elseif`s: they are converted to chains of `if/else`s.
* All expressions involving binary operators are transformed to the format we
  see in the AST, with parentheses around each operation. For example, `a+b`
  becomes `(a+b)`; and `a+b*c` becomes `((a+b)*c)`; and something stupid like
  `((((a)))/((b)))` becomes `(a/b)`.

Notice that this algorithm ignores spacing and comments. That's why you can make
this kind of change to your Procedures without causing a new version to be
created.

### Hashing Global Variables

For global variables, it's similar in concept: SHA-256 of the tokens used in the
global variable declaration (with a zero byte after each token). But again,
there are a few discrepancies from the idealized algorithm that operates
directly on the source code tokens:

* The type is always included in the hash, even if it is not explicitly present
  in the source code. This makes sure the hash will change if the type changes.
* The initializer is never taken into account. This allows changes to the
  initializer while keeping the same hash.
* Again, we use the fully-qualified name of the global variable.

## Dark Corners

* TODO: Case study: long main procedure with a hardcoded ending versus a short
  one calling and `end()` procedure; the first case will not see an updated
  ending; the second will.

* TODO: Case study: Changing the meaning of arguments between versions (even
  though the signature is unchanged) will (semantically) break the versioning.

* TODO: One big downside of this whole versioning design is that it adds a good
  deal of friction for the development of reusable libraries. The whole control
  of globals is made at the Storyworld level, so an external library cannot ever
  remove globals without this being a breaking change. Reusable libraries would
  need to release new major versions whenever they want to remove old globals
  that aren't used anymore. And old Storyworlds depending on that library would
  not be able to update without breaking compatibility with old saved states.
  Not great, but I think I can live with that for now.
