# TODO

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
