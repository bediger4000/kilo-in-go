# kilo text editor in Golang

I did [Build Your Own Text Editor](http://viewsourcecode.org/snaptoken/kilo/index.html)
only in Go instead of C. I tried to keep the spirit of `kilo`: a single
file of source code, everything as simple as possible.

I tried to check in after completing each step, but
sometimes I combined a few steps,  or fixed bugs between
steps.

# Compare C and Go

I have some experience with C, but I'm learning Go. I'd like
to use this project to compare the two languages, as well as
to internalize Go.

I found the two languages comparable in terms of expressiveness.
I transliterated the C code almost directly into Go. There was
little difference in line count in the end:

    Language  Files  Code    Comment  Blank  Total
        Go      1     797      163     86     1044
         C      1     757      152    161     1068

That's according to [sloc](http://git.bytbox.net/sloc)

## Where I found Go Better

Go's `range` operator made most iterations much simpler.

Go's standard packages had a type `bytes.Buffer` that 
replaced `struct ab` in the C version.

Go's built-in `string` type was a wash for me: C `kilo` assumes
1-byte characters, and since I was just transliterating, I did
the same in Go. I ended up using `[]byte` types for a great deal
of the code that dealt with ASCII-Nul-terminated-strings. It might
be fun to convert to Go `string` and `[]rune` and see how that
works out.

Go's memory management was a net positive. Not having to `realloc()`
all the time was easier.

Slices seem to work well, if you think of them as typed pointers-to-arrays.

The "useful zero value" for new variables means a lot less
initialization happens than in C, with a lot less room for error.

## Where I found Go Worse

C's looser idea of what comprises true and false in looping
and conditional tests lets C Kilo do some interesting things that
required extra `bool` variables in Go.

Go's Linux system call support seemed a lot less well documented,
but `kilo` does do a lot of semi-undocumented things to begin with.
Getting and setting terminal attributes, and getting into and out of
raw mode always seems a bit sketchy.

C's preprocessor macros actually would made a little clearer
code in the Go `func editorPrompt()`: I found it harder to express
"control H" in Go.

C enums worked better than Go `const`. This is a bit disingenous,
since a more idiomatic Go text editor would have several small
packages. Per-package types and constants would overcome this
objection.
