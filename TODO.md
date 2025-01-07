# TODO

* Test to add:
    * `false listen "whatever"`: I think we are trying to use `listen` as an
      infix operator in this case, which eventually panics.
        * Similar: `if false then listen "A" elseif true listen "B" end`
        * I changed `listen`'s precedence from `precPrimary` to `precNone` to
          fix it. Not sure, though, this is an area I don't really grok.
    * Expressions: literal Boolean false; string; empty string; string with line
      breaks; string with emojis and international characters.
    * For code hashing: multiple tests cases with variations of the same
      function that don't alter the hash. Like, spacing, indentation, parens
      (possibly redundant) around binary expressions, comments, escaped
      keywords. (Need more language features to make a nice one!)

Then, the remaining "unique" Romualdo features (but see also the topics after
them):

* ~~Save and restore state.~~
* Versioning.
    * Define how globals will be handled.
    * Define how to hash a procedure.
    * Implement procedure hashing.
    * ...next step for implemented versioning...
    * ~~IDEA: Try to create a visitor that reconstructs the token stream from the
      AST. Bonus points: replace all names with their FQN. This would be the
      ideal tool to hash procs and globals.~~
    * Serialize/deserialize the releases table.

Might make sense to work on these other features before (or along with) that:

* Full support for arithmetic expressions. (Because it's a nice thing to have,
  and will be relatively straightforward to bring from the old implementation,
  and will be a good thing to do if I get tired of implementing the harder
  stuff.)
* Global variables. They are also very useful and, more importantly, they are
  also versioned, so they affect versioning.
* Procedure calls. Again useful *and* related to state saving (because call
  stack).

Bug:

* Test `test/suite/expressions/bool_literal_true/src/expr.ral` fails if I remove
  the trailing `!`. Can't end Lecture with curly, it seems.
    * And more: the error is reported as an ICE (even though the error message
      is right). See also `test/suite/README.md`, which also reports a similar
      issue with an ICE.

To consider:

* Rename "space prefix" to "indentation"? Do I really need to roll my own term
  here? At least, something like "lecture indentation" if I want to be more
  specific.
* Rename "saved state" to "saved story" or something like that. Be consistent
  throughout the docs and code about this. Currently this is a bit messy, with
  different terms being used.

Optimizations:

* Procedures that don't call `listen` (must check transitively!) can't possibly
  appear on the call stack of a saved sate. So they don't have to be retained
  between versions. Maybe there's even a possibility of faster calls for those,
  since versioning is out of the table.

Older TODOs (review):

* Consider compressing strings and lectures on the bytecode. Complicates the VM,
  but may be worth it. Could use a simple algorithm like
  [shoco](https://ed-von-schleck.github.io/shoco/), or gzip. Again, not super
  happy about complicating the VM (but then, GDScript can gzip. Can "normal
  browser Javascript" do it, too without extra dependencies?). Also to consider:
  someone wanting smaller size could compress the whole compiled Storyworld at
  once, right? Even better compression, and more control to users.
* Eventually, will need an array of `meta`s: one per chunk, with the `meta`s
  defined for each chunk. Or maybe this could be a map, because most
  procedures will not have `meta`s.
    * Actually, I am confused. Need to know what is part of the static CSW
      and what is dynamic state the VM maintains.
