# The Romualdo Language Specification

***Warning:** This is all tentative, incomplete, and work-in-progress!*

(That said, the things I am placing here (as opposed to `design.md`) are the
ones which are *a bit* more solidified in the design.)

## Preliminaries

Let us start with some higher-level concepts that don't fit too well into the
notion of a formal grammar.

### Storyworlds

The Romualdo language is designed for creating Interactive Storytelling
experiences. You could probably used it for some "general programming" tasks,
but in that case you'd be better served by one of the many languages designed to
be general-purpose languages.

To emphasize that Romualdo is not really a general-purpose programming language,
we don't say "Romualdo programs" or "Romualdo apps"; we say "Romualdo
**Storyworlds**".

### Stories

The result of running a Storyworld is a **Story**.

### Driver Program

The Romualdo tools are designed so that a Storyworld can be run as part of a
program written in another programming language (as long as we have a Romualdo
Virtual Machine available for that language).

That program that "hosts" the Romualdo Storyworld is called the **Driver
Program**.

### Packages

A Storyworld is composed of one or more Romualdo source files (usually more than
one for any realistic Storyworld). You can group the source files into
**Packages** for better modularity, organization and maybe reusability.

A Package corresponds to a directory in the filesystem. All Romualdo source
files in a single directory are part of the same Package.

Packages form a hierarchical structure of Packages and **Subpackages**. The fact
that a given Package is a Subpackage of another one is usually irrelevant, so we
normally simply call everything a Package. (Except when we want to emphasize the
hierarchical relationship for some reason.)

The Package at the root directory of a Storyworld as a whole is called the
**Root Package**. Other Packages have names matching the directory name where
they reside. The full name of a Package is like a directory path. Something like
`/package/subpackage/my_package`.

A Package name must be a valid Romualdo identifier. It can't be `std`, though,
which is reserved for the Romualdo standard library.

Restricting Package and file names to the ASCII character set can avoid
headaches (e.g., when exchanging files with other people), but Romualdo doesn't
enforce this.

### The `main` Procedure

The execution of a Storyworld starts from a Procedure called `main` located at
the Root Package (in other words, the entry point of a Storyworld has a Fully
Qualified name of `/main`).

## Source file

As we saw above, a Storyworld is composed of a set of Romualdo source files that
happen to be organized in a hierarchy of Packages. Romualdo can parse each each
of these files independently, so the Romualdo source file is the top-level rule
of our grammar.

A source file is simply a sequence of Package imports followed by a sequence of
declarations:

```ebnf
sourceFile = packageImport* declaration* EOF ;
```

All Romualdo source files must be encoded in UTF-8.

## Package imports

We use package imports to make symbols declared in other Packages available to
the current source file.

```ebnf
packageImport = "import" packagePath [ "as" IDENTIFIER ] ;

packagePath = [ "/" ] packageSegment ( "/" packageSegment )* ;

packageSegment = ".." | IDENTIFIER ;
```

Import paths can be absolute or relative to the current Package:

```romualdo
import /util/names        \# Absolute: looks for utils/names at the Root Package
import encounters/random  \# Relative path: looks for encounters/random at the current package
```

The example above have paths with length 2, but any positive length is fine:

```romualdo
import /utils
import some/very/deeply/nested/package
```

Just like in filesystem paths, you can use `..` to refer to the parent Package
of the current Subpackage.

```romualdo
import ../utils
import ../
```

You should think twice before using `..` , though. It quickly gets confusing,
especially if you decide to reorganize your Package hierarchy.

### The `std` Package

A special, magic case of Package imports is the Romualdo standard library. It is
always available as `std` without the need of importing it.

### Accessing symbols from imported Packages

By default, imported symbols are available as `package_name.Symbol_name`:

```romualdo
import /util/random

\# Later on, assuming there's a GetRandomName() Procedure declared
\# in the /util/random Package:
name = random.GetRandomName()
```

You can **rename a Package import**, too. It is necessary when you are importing
multiple Packages that have the same name, or if the Package name is a Romualdo
reserved word:

```romualdo
import /util/random as ur
import /tools/random as tr
import my_poorly_named_packages/passage as myPassage

\# Later on:
name = ur.GetRandomName()
```

### Package import Errors

Let's see some possible errors when importing Packages.

You can't use `..` to go beyond the Root Package:

```romualdo
import /..         \# Error! Cannot go beyond the Root Package.
import /foo/../..  \# Error! For the same reason.
```

You can't import the Root Package:

```romualdo
import /my_package/..  \# Error! Cannot import the Root Package.
import /               \# Error! The grammar itself forbids this case.
```

### What is imported?

Only symbols whose names start with an uppercase letter are imported. By
"uppercase letter", we specifically mean Unicode code points assigned to the
"Lu" ("Letter, uppercase") category.

All other symbols are visible only within the Package they are declared.

## Types

Romualdo is strongly-typed. There are one or two corners where typing gets a bit
unsafe, but generally types are very precise and clear.

```ebnf
type = "void"
     | "bool"
     | "int"
     | "float"
     | "bnum"
     | "string"
     | "[" "]" type
     | "map"
     | procedureType
     | userDefinedType ;

procedureType = ( "function" | "passage" ) "(" [ typeList ] ")" ":" type ;

typeList = type ( "," type )* ;

userDefinedType = IDENTIFIER [ "." IDENTIFIER ] ;
```

The supported types are:

* `void`: A non-type. Used when a type is formally required, but is not really
  needed (like the return value of a function that doesn't return anything).
* `bool`: Booleans, true or false, no surprise here, right? (Default value:
  false)
* `int`: A signed integer number, no less than 32 bits. You shouldn't really
  count on the size (no pun intended). (Default value: false)
* `float`: A floating-point number, most likely a IEEE 754 binary64 (double
  precision) number (but, again, you should not count on that). (Default value:
  NaN)
* `bnum`: Chris Crawford's bounded numbers, which I hope will be nice for doing
  story things like character models (that's what `bnum`s were designed for,
  anyway).
* `string`: A string of characters, meant to hold UTF-8-encoded text.
* Array: A sequence of zero or more elements of the same type. `[]int` is an
  array of `int`s, `[]string` is an array of `string`s, and so on.
* `map`: An associative array mapping string keys to values of any other type.
  While all keys must be strings, the values can be of any, possibly mixed
  types. The fact that a `map` value can be of any type is the reason for the
  existence of type-unsafe corners of the language. It's also handy for
  communication with the Driver Program, as many modern programming languages
  have types that are a superset of what a Romualdo `map` is.
* Procedures: Procedures taking a certain set of parameters and returning a
  certain type.
* User-defined types: Those can be declared in the same Package or in some other
  Package and imported, so `userDefinedType` allows for things like `myType` or
  like `thatPackage.ThatType`.

TODO: Point to the (yet-to-be written) section in which we explain how we deal
with the `map` type unsafeness. (Thinking of default values on "soft errors".)

TODO: Need some thinking about how NaNs are handled. May want to leave this open
("don't count on this"), as being specific here can lead to unnecessary
difficulties in implementing VMs in certain languages.

## Declarations

```ebnf
declaration = globalsBlock
            | functionDecl
            | passageDecl ;
```

TODO: User-defined types: type `alias`es and `struct`s (`class`es?).

### Globals

Global variables must be declared in a `globals` block.

```ebnf
globalsBlock = "globals" [ "@" INTEGER ]
               varDecl*
               "end" ;

varDecl = IDENTIFIER [ ":" type ] [ "=" expression ] ;
```

Each `globals` block has a version (if omitted, it is assumed to be 1). There
can be only one `globals` block of any given version in any given Package.

A `globals` block of a latter version must redeclare all globals declared in the
previous version, using the same type as before. The initialization expression
can be different, though: it will be used instead of the old one only when
starting a new Story.

Of course, it also valid to add new globals to a new version of a `globals`
block.

```romualdo
\# No explicit version provided, so 1 is used.
globals
    EndGame: bool = false
    artifactsCount: int = 0
end

\# Here we are explicitly saying this version 2.
globals@2
    EndGame: bool = false              \# Fine, same as version 1
    artifactsCount: int = 1            \# Fine, just initialization changed
    favoriteColor: string = "blue"     \# Fine, a brand new global variable
end

\# One more version!
globals@3
    EndGame: string = "sure!"          \# Error! Changed the type
                                       \# Error! Didn't redeclare all globals
end
```

The initialization expression is optional. If omitted, each variable is
initialized by the default value of it's corresponding type.

The type can be omitted if it can be inferred from the initialization
expression:

```romualdo
globals
    EndGame = false      \# Fine, `EndGame` is a bool because `false` is a bool
    artifactsCount       \# Error! Type not informed and can't be inferred
end
```

*Note:* We don't allow removing or changing types of globals from one version to
another because this could potentially break ongoing Stories. And we require all
variables to be redeclared to avoid having globals confusingly scattered over
different `globals` blocks.

### Procedures

Procedures are where things happen. Romualdo supports two types of procedures:
Functions and Passages. They are completely equivalent in terms of capabilities,
but each of them supports a different syntax, which is more appropriate for
certain things.

**Functions** are geared towards traditional programming. They are the obvious
choice for the cases in which you are implementing the brains of your
Storyworld. For example, maybe you want to have some sort of simulation to
generate certain events for an ongoing Story: you'd want to use functions to
implement this simulation.

```ebnf
functionDecl =  "function" [ "@" INTEGER ] IDENTIFIER "(" [ parameters ] ")" ":" type
                statement*
                "end" ;

parameters = parameter ( "," parameter )* ;

parameter = IDENTIFIER ":" type ;
```

**Passages** are ideal for saying to the Player of your Storyworld. Typically,
when you are effectively *telling* the Story, you'll want to use Passages.

```ebnf
passageDecl =  "passage" [ "@" INTEGER ] IDENTIFIER "(" [ parameters ] ")" ":" type
               LECTURE
               "end" ;
```

To reiterate: a Function can do anything a Passage can do and vice-versa. The
choice is a matter of convenience, and the difference is that the body of
Function is sequence of statements, while the body of a Passage is what we call
a Lecture. TODO: Point to the section in which we describe Lectures.

### Statements

Statements are language constructs that do stuff. They don't have a value.

```ebnf
statement = varDeclStmt
          | assignmentStmt
          | blockStmt
          | whileStmt
          | ifStmt
          | returnStmt
          | sayStmt
          | expression ;

varDeclStmt = "var" varDecl ;

assignment = [ call "." ] IDENTIFIER "=" expression ;

blockStmt = "do"
            statement*
            "end" ;

whileStmt = "while" expression "do"
            statement*
            "end" ;

ifStmt = "if" expression "then" statement*
         elseif*
         [ "else" statement* ]
         "end" ;

elseif = "elseif" expression "then" statement*

returnStmt = "return" [ expression ] ;

sayStmt = "say" LECTURE "end" ;
```

Some notes about the statements:

* Local variable declarations can appear anywhere (anywhere a statement can
  appear, *bien entendu*). A local variable exists from the point it is declared
  until the end of its scope. A local variable cannot shadow an existing local
  variable.
* The only purpose of `do`...`end` statements is to create blocks, which form a
  scope and therefore allow to control the lifetime of the enclosed local
  variables. TODO: I honestly didn't intend to have this on the language, but I
  added them to allow me having local variables before I have other
  block-defining statements. Maybe I'll remove it in the future.
* Nothing surprising about `while` loops: execute a sequence of statements as
  long as a given expression evaluates to `true`.
* Nothing surprising with `if`s either.
* Ditto for `return`s.
* The `say` statement is used to send information to the Driver Program that is
  running the Storyworld. Typically, it is used to describe events that happened
  in the story and need to be somehow shown to the player (the *how* in the
  *somehow* is responsibility of the Driver Program, not of Romualdo). TODO:
  Link to the description of Lectures.
* Expressions can be used as statements. Depending on the expression this can be
  useful (a function call is often used for its side-effects only) or useless
  (an expression like `1 + 1` by itself serves no purpose -- but is considered
  valid nevertheless).
* TODO: `for` loops are the most notable absence here. I want to support things
  like `for i in range(0, 10) do ... end` and `for t in arrayOfThings do ...
  end`, but then I'd have to store some additional state (the `range()` result,
  the current pointer into `arrayOfThings`) somewhere and don't know where this
  somewhere would be. Probably in a local variable. Anyway, for now, `for` loops
  are not available at all.
    * TODO: We can probably go with two versions of `for`: one for counting,
      another for iterating over collections (maybe `for` and `foreach`?).

## Expressions

Expressions evaluate to a value. The different levels of precedence are encoded
in the grammar (which makes the grammar weirder to look at, but will hopefully
translate more directly to the implementation).

```ebnf
expression = logicOr ;

logicOr = logicAnd ( "or" logicAnd )* ;

logicAnd = equality ( "and" equality )* ;

equality = comparison ( ( "!=" | "==" ) comparison )* ;

comparison = addition ( ( ">" | ">=" | "<" | "<=" ) addition )* ;

addition = multiplication ( ( "-" | "+" ) multiplication )* ;

multiplication = exponentiation ( ( "/" | "*" ) exponentiation )* ;

exponentiation = unary ( "^" exponentiation )* ;

unary = ( "not" | "-" | "+" ) unary
      | call ;

call = "listen" expression
     | primary ( "(" [ arguments ] ")"
               | "." IDENTIFIER
               | "[" expression "]"
               )* ;

arguments = expression ( "," expression )* ;

primary = "true"
        | "false"
        | FLOAT
        | INTEGER
        | STRING
        | arrayLiteral
        | mapLiteral
        | "(" expression ")" ;

arrayLiteral = "[" [ expression ( "," expression )* [ "," ] ] "]" ;

mapLiteral = "{" [ mapEntry   ( "," mapEntry   )* [ "," ] ] "}" ;

mapEntry = ( IDENTIFIER | STRING ) "=" expression ;
```

Notes about expressions:

* The `listen` expression is used to get input from the player. Its `expression`
  argument must evaluate to a `map`, which represents the alternatives the
  player has. `listen` transfers the control to the Driver Program, which can
  access the data from this `map`, show alternatives to the Player, get a choice
  from the Player and give the control back to the Storyworld, passing to it the
  Player choice. The Player choice is the value of the `listen`, and is
  always a `map`.
    * TODO: The initial implementation of `listen`  will take a `[]string`
      argument instead of `map` and return an `int`.
* Logical operators `and` and `or` have short-circuited evaluation.
* Note the syntax for literal arrays and maps. Trailing comma allowed.
* TODO: blend for bnum!

## Versioning

**TODO**, but in summary:

* The problem we are solving here is allowing to upgrade or patch a Storyworld
  without breaking saved ongoing stories.
* Every Procedure and `globals` block has a version.
* If omitted, version is implicitly assumed to be 1.
* Things cannot be changed between releases of a Storyworld, only added.
* So, it's OK to add a new version of a Procedure or `globals` block.
* All new stuff added to a Package in a release must have the same version, that
  is one higher than the previously highest version.
* This may generate holes in versioning, but that's fine. It's assumed that
  those are the same as the highest defined version that is lower than the
  missing version. (Shouldn't make diference to the backend; it's a check made
  on the frontend to make sure things are a bit easier to understand, because
  versions will match with a release *within a Package*.)
* We can call a specific version of a Procedure: `proc@3()`. Omitting the
  version causes the latest version to be called.
* See also the descriptions of `globals` blocks and Procedures for the details
  on versioning specific to them.
