# TODO

* More test kinds to add:
    * Tests cases for errors (is it failing as expected for wrong code?)
    * Someday: tests checking the exit code of the `romualdo` tool.
        * Like, tests with syntax errors and such.

* I need either of these to write interesting tests:
    * Explicit `say` statements in `functions`.
    * Backslashed statements in Lectures, plus support for `{1+1*3}` to allow me
      to output arbitrary expressions.

Would be nice to have both! 😉

* Test to add:
    * `false listen "whatever"`: I think we are trying to use `listen` as an
      infix operator in this case, which eventually panics.
        * Similar: `if false then listen "A" elseif true listen "B" end`
        * I changed `listen`'s precedence from `precPrimary` to `precNone` to
          fix it. Not sure, though, this is an area I don't really grok.

Then, the remaining "unique" Romualdo features (but see also the topics after
them):

* Save and restore state.
* Versioning.

Might make sense to work on these other features before (or along with) that:

* Full support for arithmetic expressions. (Because it's a nice thing to have,
  and will be relatively straightforward to bring from the old implementation,
  and will be a good thing to do if I get tired of implementing the harder
  stuff.)
* Global variables. They are also very useful and, more importantly, they are
  also versioned, so they affect versioning.
* Procedure calls. Again useful *and* related to state saving (because call
  stack).

Older TODOs (review):

* Eventually, will need an array of `meta`s: one per chunk, with the `meta`s
  defined for each chunk. Or maybe this could be a map, because most
  procedures will not have `meta`s.
    * Actually, I am confused. Need to know what is part of the static CSW
      and what is dynamic state the VM maintains.
