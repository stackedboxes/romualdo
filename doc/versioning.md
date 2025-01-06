# The Romualdo Language Specification

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
may affect how the user "perceives" versioning. TODO: Explore this, add
examples! Like, a long main procedure with a hardcoded ending versus a short one
calling and `end()` procedure; the first case will not see an updated ending;
the second will.)

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
default values (if I ever support that). And (nasty trap!): completely changing
the meaning of arguments in a new version of a procedure will break the semantic
of saved states!

**TODO:** Hey, can't globals and procedures be internally represented by their
hash only? (With the debug info file providing a mapping to user-friendly
names.) *Counterpoint:* I will want to have some way to filter/select Passages
based on signature and meta/static variables; would it be interesting to filter
also by name? *Stronger counterpoint:* What would I gain from this? Better to
have them internally represented by an index into an array of
procedures/globals, which is simpler and faster to access.

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

Since this is all based on hashes of the actual code of the procedures and
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
        * NEW COMMENT, 2024-11-29. But a lexeme referring to a name
          (`package.name`) may be ambiguous, because different packages can be
          imported such that they are referred to by the same name! So, I guess
          that for hashing purposes, all names should be the FQN.
            * The scanner doesn't know about those, though! Looking like an
              AST-based approach would be better here, too!
    * Add a zero byte to the hash computation (this is a single byte with all
      bits set to zero, not a string with an ASCII "0" character!).
        * This zero byte is there to disambiguate between two consecutive tokens
          that could be interpreted as a single different token. For example,
          this makes sure tokens `else` and `if` are hashed to a different value
          than the single token `elseif`. (My implementation of `codeHasher`
          always generates separate "else" and "if" tokens, so this example is
          more theoretical than practical!)

**Notes from my `codeHasher` implementation:**

* We always generate separate "else" and "if" tokens (with the corresponding
  "end" tokens for each "if").
* Binary operators always emit "(" before and ")" after them, so that the right
  precedence is maintained. Notice that the parentheses in the source code are
  not represented in the AST, so we need to do that when reconstructing the
  source code for hashing. A side effect is that removing or adding redundant
  parentheses to the source code does not change the hash (quite nice if works
  as I hope; kinda scary, too).

Note that by taking into account only the tokens, we allow changes to formatting
and comments (which do not affect the generated code).

MD5 should be a good choice here. It should be fast enough and security is not a
concern here. (**IDK**, probably going with SHA-256 just for ultra-paranoia
about collisions).

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
(at scanner level we don't know about inferred types).

**TODO:** I hate the difference in handling of Procedures and globals.

*Idea:* What if globals were also handled in the scanner. I would need to forbid
any changes to globals, including adding or changing an initializer. Would this
be OK? One can always have a custom initialization function to "re-initialize"
the global at Story start. *Or,* make the type required in all global
declarations (no inference for globals); a bit annoying, but then we'd always
have access to the type at scanner level, so we could always get the hash right.

**Except...** the hash must use the var FQN! Otherwise there may be clashes. In
fact, same for procedures, right? If I move a procedure to a different Package,
the hash must change! Or does it?

**So...** maybe two stages? "Partial" hash in the scanner, then a second step
triggered from the parser e concatenate this partial hash with the package name
and re-hash -- and *this* is the final hash. Could work, but *so* convoluted!
AST looking like the simpler way...




## Dark Corners

* TODO: Case study: long main procedure with a hardcoded ending versus a short
one calling and `end()` procedure; the first case will not see an updated
ending; the second will.





**TODO**, but in summary:

* The problem we are solving here is allowing to upgrade or patch a Storyworld
* Things cannot be changed between releases of a Storyworld, only added.
* So, it's OK to add a new version of a Procedure or a new global variable.
* See `design.md` for details on the current (still evolving) design.
