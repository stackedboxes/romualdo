# The Romualdo Language

*A programming language for Interactive Storytelling.*

## Status

Very not suitable for anything. Ongoing design and development. This repo is a
language redesign of a [previous, unfinished
incarnation](https://github.com/lmbarros/romualdo-language) of Romualdo.

## A brief note on the vision

This project is very experimental, and I myself am not entirely sure where it
will head to. Anyway, let me give you *some* clue of what I am trying to achieve
here. (And to be clear, this is the vision, not the current status!)

Essentially, this not very much unlike [Ink](https://www.inklestudios.com/ink/):
you can think of Romualdo as a story-engine you can embed in a game written in
some other programming language.

Ink is really nice, but I believe it leans too much to the "writing tool" end of
the "writing tool-programming language" spectrum. Nothing wrong with that, but I
believe that making it easier to write "real code" will open some nice
opportunities for interactive storytelling (or however you want to call it).

The code you write should look more or less like this simplistic example:

```romualdo
passage AtTheLandmark(landmark: string): void
    if landmark == STONE_CIRCLE and moonPhase == FULL then
        say
            The druids were there and killed you.
        end
        return
    end

    say
        You are at the {landmark}, wondering what to do next.

        More text, more text, and even more text.
    end

    choice = listen("What do you do?")

    if choice == "wait" then
        WaitAtTheLandmark(landmark)
    else
        UpdateTheWorldStateInSomeComplexFashion()
        DoSomethingElseAtTheLandmark(landmark)
    end
end
```

`say` is the way to send data from Romualdo to its host program, and `listen` is
the other way around. Think of them as you having a conversation with the
player.

Here I used just plain text to communicate between the Romualdo storyworld and
the host program. We should be able to use some structured format like JSON or
XML for easier integration with your game or engine.

## Credits

* The Romualdo Language syntax is in no small extent inspired by
  [Lua](http://www.lua.org).
* [Ink](https://www.inklestudios.com/ink/) was a considerable influence, too.
  Romualdo's goals are quite different (it is meant to more like a traditional
  programming language), but learning about Ink and using it a little bit lead
  to a large Romualdo redesign.
* The implementation is strongly based on Bob Nystrom's excellent [Crafting
  Interpreters](http://www.craftinginterpreters.com). I cannot understate how
  this book was useful.
