# TODO

* Implement `if`.
* Add tests for interactivity.

At this point we can make interactive stuff.

* Test to add:
    * `false listen "whatever"`: I think we are trying to use `listen` as an
      infix operator in this case, which eventually panics.
        * And maybe the solution is just to change `listen`'s precedence from
          `precPrimary` to `precNone`. Not sure, though, this is an area I don't
          really grok.

Older TODOs (review):

* Tests cases for errors (is it failing as expected for wrong code?)
* Eventually, will need an array of `meta`s: one per chunk, with the `meta`s
  defined for each chunk. Or maybe this could be a map, because most
  procedures will not have `meta`s.
    * Actually, I am confused. Need to know what is part of the static CSW
      and what is dynamic state the VM maintains.
* Someday: tests checking the exit code of the `romualdo` tool.
    * Like, tests with syntax errors and such.
