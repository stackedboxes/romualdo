# TODO

Next up:

* Implement deserialization for DebugInfo.
* Implement a `run` command that executes a CompiledStoryWorld saved to disk.
* Implement a `--trace` flag to `run`
* Implement a `disassemble` command.
    * Add flags allowing to just print a summary (or "index").
    * Allow show details about just one (or some) selected things from the
      "index".

Once I have those in place I have pretty much the whole workflow working. (Maybe
missing only the "build new version" flow, which will allow to create new
versions of Storyworlds)

Then, the way forward is:

* Make testing work with both `walk` and `run`.
* Implement `listen`
* Implement `if`.
* Add tests for interactivity.

At this point we can make interactive stuff.

Then, rethink file extensions. Maybe:

* `.ral` for sources
* `.ras` for compiled storyworlds
* `.rad` for debug info
* `.raf` for saved ("frozen") state (doesn't really need to be standardized)

Older TODOs:

* Tests cases for errors (is it failing as expected for wrong code?)
* Eventually, will need an array of `meta`s: one per chunk, with the `meta`s
  defined for each chunk. Or maybe this could be a map, because most
  procedures will not have `meta`s.
    * Actually, I am confused. Need to know what is part of the static CSW
      and what is dynamic state the VM maintains.
* Someday: tests checking the exit code of the `romualdo` tool.
