# Language Design Notes

First of all, here are the [specs of the previous Romualdo
iteration](https://github.com/lmbarros/romualdo-language/blob/master/doc/maybe-the-ideal-grammar.md).

For this iteration, the main changes are:

* The output from Romualdo to the host program is text only. At least initially.
  And, of course this text could be structured and interpreted by the host in
  various ways -- but from Romualdo's perspective, it's just text.
* Outputting text is much simpler. Basically, within a `passage` things are
  scanned differently: everything is text that will be output. Programming
  constructs are still possible, but keywords must be prefixed by a backslash
  (`\`).

## Passages

Tentative example:

```romualdo
passage thePassage(): void

    This is text that is outputted when running the passage. The common
    indentation shared by all lines is removed.

    \if whatever then
        And this is outputted conditionally. BTW, backlashed keywords are always
        valid, but we may not need them always, like shown here. We *know* that
        an expression and then `then` must come after `\if`. Now, maybe a bad
        idea but an `end` with less indentation can be interpreted as an `end`
        instead of text with invalid indentation.
    end
end
```

## Passages x Functions

In principle, both should be allowed to do the same things. It's just that the
syntax accepted by each one is different, favoring either text or code.

## Principles

Rough and kinda conflicting, but these are some principles I am trying to
follow. These are more about implementation than design.

* **Maintainability over performance.** At least within reasonable limits. I am
  creating this because I want to use it, not because I want to maintain it.
* **User friendliness over best practices.** I don't mind having two almost
  identical functions if they can provide better error messages (compared with
  merging them into a single function). "User" in this case means "me".
