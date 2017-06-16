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

Using [sloccount](???)

CKilo: 971 lines of ANSI C
GoKilo: sloccount doesn't work

## Better

Go's `range` operator made most iterations much simpler.

Go's standard packages had a type `bytes.Buffer` that 
replaced `struct ab` in the C version.

## Worse

## Just Different
