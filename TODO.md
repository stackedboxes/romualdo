# TODO

* Test to add:
    * `false listen "whatever"`: I think we are trying to use `listen` as an
      infix operator in this case, which eventually panics.
        * Similar: `if false then listen "A" elseif true listen "B" end`
        * I changed `listen`'s precedence from `precPrimary` to `precNone` to
          fix it. Not sure, though, this is an area I don't really grok.

Then, the remaining "unique" Romualdo features (but see also the topics after
them):

* ~~Save and restore state.~~
* Versioning.
    * Add some unique ID to both Storyworlds and saved states. This would be a
      crude way to detect if a saved state is compatible with the current
      Storyworld (the ID must match and the versions must be compatible). Not
      great, but better than nothing.
        * Can we do better than this? Well, at state loading we could check if
          all procedures/versions at the call stack exist in the Storyworld.
          Same for global variables. And heck, even for local variables for the
          procedures/versions on the call stack.

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
