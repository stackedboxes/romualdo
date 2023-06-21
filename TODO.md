# TODO

Next up:

* Implement serialization/deserialization for CompiledStoryWorld and DebugInfo.
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

* ~~Add package names to parsed stuff.~~
* Tests cases for errors (is it failing as expected for wrong code?)
* Add the bytecode compiler and VM for what I already have.
    * ~~One chunk per procedure~~
    * ~~Constants in the CSW~~
    * ~~Mapping between global names and indices go to a separate debug info file.~~
    * ~~So, CSW has an array of chunks.~~
    * ~~And a field with the index of the latest version of `main`. To know where
      to start.~~
    * Eventually, will need an array of `meta`s: one per chunk, with the `meta`s
      defined for each chunk. Or maybe this could be a map, because most
      procedures will not have `meta`s.
        * Actually, I am confused. Need to know what is part of the static CSW
          and what is dynamic state the VM maintains.
* Someday: tests checking the exit code of the `romualdo` tool.
