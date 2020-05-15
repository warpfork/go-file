go-file
=======

Yet Another file abstraction library for golang.

Why
---

1. The golang stdlib doesn't provide any interface abstractions over files.  It's not possible to do an in-memory FS, for example, which you might want to do for writing faster unit tests.
2. Sometimes it's important to write programs which handle posix metadata in a clear, consistent, _host-platform agnostic_ way.  This requires metadata structures that have no host-specfic depedencies.
  (The stdlib `os` package has almost the opposite design goals -- its aim is to get things done on your physical host, and it's not shy about exposing platform details.)
3. None of the other third party libraries seem to have nailed it yet.

There are, yes, lots of other libraries that have approached this problem.

The major differences between go-file and other libraries are:

- we're embracing POSIX and most of its details (such as uid, gid, etc) -- whereas many other libraries try to smooth over those properties and pretend they don't exist.
- we're leaning into the Go type system to make as many errors compile-time impossible as we can (most essential types have constructors that enforce correctness of parameters, rather than shrugging and allowing create-by-casting).
- a major intention of this library is to make it possible to reason about posix-style metadata of files, _regardless_ of what system you compiled on (e.g., the `syscall` package is *not* acceptable for any of our public APIs).

(And yes, if a cast that peeks into the `syscall` package is sufficiently commonly used by users to get at some important detail,
we consider that still _in effect_ a defacto appearance of the `syscall` package in the "public API".
That means the `os.FileInfo.Sys().(*syscall.Stat_t).Gid` hack?  Nope.  Not here.  Not in this library.)

There are also some more itemized reasons that various other libraries have been considered but rejected before we started a new one... jump to the [Alternatives](#alternatives) section if you'd like to see those.


Usage
-----

```go
package main

import (
	"github.com/warpfork/go-file/file"
)

func main() {
	// TODO :)  so sorry, check back soon
}
```


Alternatives
------------

### Why not 'billy'?

https://github.com/go-git/go-billy

Billy is sorta fine for what it does.  And they did a pretty good job at abstraction in general.
It definitely works well for what it was written for: which is to support tests (especially in-memory filesystems) for the go-git project.

A couple qualms:

1. It uses the `os.FileMode` abstraction, which...
	- is not my favorite part of the `os` package.
		- It's too prone to bitmath, and not all the bits are what you'd expect;
		- it's hard to use it to help enforce whole-program correctness since you can cast to it too easily;
		- and it's just plain irritating to use as a bonus: the check for "is this a plain file", _the most common thing you'll ever check_, gets bitmathy (you check that a mask of it is equal to zero).  Ick.
	- means you import the whole `os` package.
		- Honestly, maybe this isn't that big of a deal in any practical sense.  But it still makes me feel... dirty.

2. It uses the `os.FileInfo` abstraction, which...
	- repeats the qualm about importing the whole `os` package.
	- is _not **enough**_.
		- `os.FileInfo` doesn't include uid or gid, for example.  Many applications I work on require that info.
		- While you can still get the relevant info through the `fi.Sys() interface{}` and then cast it to `syscall.Stat_t`...
			- and _techinically_ all the fields on that type are exported, so as a third-party author you can create and wield values of it...
				- all this is getting deeply into "bandaid that's technically (if barely) correct" territory, and I don't like to be anywhere near that territory.
				- none of this helps us if part of our goal is to reason about those metadata safely _regardless of what platform we're on_ -- the syscall package _does change_ depending on the build platform.

3. The versioning of this package (it changes import path somewhat often) has been irritating.
	- Personal opinion.  I know some people like the major-version-in-import-path thing.  I don't.  It's caused me to waste much time.

go-billy is overall pretty excellent.  If it suits you, use it!
It's just also not what we needed when we wrote go-file.

### Why not 'net/http.FileSystem'?

https://godoc.org/net/http#FileSystem

The `net/http` package has a brief filesystem interface.  It _does_ make it possible to generate in-memory representations of filesystems.

However, it's a very (very) minimal interface.  It doesn't even come close to satisfying our needs around thorough handling of POSIX metadata, etc.

This is okay!  It's no insult to that package.  It just does exactly what the `net/http` package needed, and no more.  This is good design in action.
It's just also not what we needed when we wrote go-file.

### Why not 'afero'?

https://github.com/spf13/afero

Afero is a filesystem abstraction library.

1. It uses the `os.FileMode` abstraction, which...
	- Same problems as when 'billy' used it.  See above.

2. It uses the `os.FileInfo` abstraction, which...
	- Same problems as when 'billy' used it.  See above.

3. I'm just not a fan of its kitchen-sink design.
	- Functions like 'NeuterAccents' for mutating strings do not belong in the root package of a library like this (if they belong anywhere in it at all).
	- All the implementations of the interface are in the root package together.  It's unclear without deep investigation if writing your own implementation of the interface from outside the main package is practical; if the author didn't, why should I assume a third party can?
	- Just... look at the godoc and compare it to go-billy.  It's... there's a difference.

4. It contains a direct dependency on the `syscall` package.
	- Just... No.

If use of the `os.FileMode` and `os.FileInfo` abstractions was acceptable, I'd much prefer to use the billy library, due to its overall focus on coherent abstraction.
Given that it's not, afero is doubly out the window, for having the same limitations, and a lot more.
