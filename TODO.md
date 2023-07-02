# TODO

Next up:

* Rename "first chunk" to "initial chunk"?
* File names are still not correctly recorded. Here's what I get from
  disassemble:  
     First procedure: 0 [main(), test/hello_escaped/]

Once I have those in place I have pretty much the whole workflow working. (Maybe
missing only the "build new version" flow, which will allow to create new
versions of Storyworlds)

Then, the way forward is:

* Make testing work with both `walk` and `run`.
* Implement `listen`
* Implement `if`.
* Add tests for interactivity.

At this point we can make interactive stuff.

Older TODOs:

* Tests cases for errors (is it failing as expected for wrong code?)
* Eventually, will need an array of `meta`s: one per chunk, with the `meta`s
  defined for each chunk. Or maybe this could be a map, because most
  procedures will not have `meta`s.
    * Actually, I am confused. Need to know what is part of the static CSW
      and what is dynamic state the VM maintains.
* Someday: tests checking the exit code of the `romualdo` tool.
