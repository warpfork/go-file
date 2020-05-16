package file

import (
	"io"
	"time"
)

// Cabinet is access to a swatch of a filesystem.
//
// Cabinet can be backed by a real (presumably posix-ish) filesystem,
// or backed by an in-memory implementation,
// or backed by a read-only view of a tar file,
// or... whatever else you can dream of that fits the interface!
//
// Fair warning: if a Cabinet instance is the only thing that accesses
// whatever its backing storage is, then sure, you can assume no conflicting
// mutations between serial accesses to the Cabinet and any Handles yielded
// from it, or mutex access to these things, etc.
// But if the Cabinet instance is backed by some storage which can be
// accessed in more than one way -- and this is the _common case_ --
// e.g. any Cabinet backed by a posix filesystem is this! -- then it's
// impossible to assure a lack of concurrent modifications from within
// this interface alone.  You'd have to establish contracts based on whatever
// the backing store is, and enforce it using mechanisms that work across
// all possible ways of accessing that backing store.
//
// REVIEW: name: file.Tree?  file.Root?  file.Subtree?
// Want to avoid connotations of "mount", or "drive" -- isn't necessarily either;
// doesn't necessarily hold any of the data (like total size, for a random example) you'd expect to be able to get from such a thing.
//
// REVIEW: Is it even distinct from Handle?  Perhaps not!
// Maybe we'll end up with some things that are best done with interface detection?
// Maybe being able to ask what cabinet (root, tree, forest, whatever) a handle belongs to will be useful?
type Cabinet interface {
	OpenRoot() (Handle, error)

	// REVIEW: dislike putting this in a "global" place, but does it really make sense to put it on Dirs or anything else?
	//   The function would have to error if the destination was in a different cabinet.  Or the path resolution ... handles would have to keep.... oioioi.
	//   And it doesn't apply to an _open_ handle.  (This... is a whole Thing to follow up on.  The kernel in linux certainly does have variations of calling the syscalls with fds.  It's the golang os package that drops this, afaict.)
	//
	// REVIEW: maybe putting this on cabinet is the error... and what we should pursue is having something that can act a bit like a cursor, but only on paths (no open handles).
	Rename(old Path, dest Path) error
}

type Handle interface {
	Kind() Kind
	ReadMetadata(*Metadata) *Metadata // Reads metadata -- if a pointer is given as a param, it will be modified and the same pointer returned; or, if given nil, will return a newly allocated structure.
	File() File                       // File specializes the value into the File interface, or panics if the Kind is incorrect.  (You could also do this by casting, but this method provides a better error.)
	Dir() Dir                         // Dir specializes the value into a Dir, or panics if the Kind is incorrect.  (You could also do this by casting, but this method provides a better error.)
	io.Closer                         // Close is an operation defined on all kinds of handle.

	Xattrs() Xattrs // Returns an interface that can be used to access "extended attributes".
}

type (
	// Kind is an enum.
	// Its members are the `Kind_*` types.
	Kind interface {
		_Kind()
		String() string
		GoString() string
	}

	Kind_File        struct{}
	Kind_Dir         struct{}
	Kind_Symlink     struct{}
	Kind_Fifo        struct{}
	Kind_Socket      struct{}
	Kind_BlockDevice struct{}
	Kind_CharDevice  struct{}

	// Note that "hardlink" is not a kind of file.
	//  Hardlinks are a contextual concept, and they don't make any sense outside of the scope of a single "cabinet" (in posix: a mount), which makes their semantics very wobbly to say the least.
)

func (Kind_File) _Kind()        {}
func (Kind_Dir) _Kind()         {}
func (Kind_Symlink) _Kind()     {}
func (Kind_Fifo) _Kind()        {}
func (Kind_Socket) _Kind()      {}
func (Kind_BlockDevice) _Kind() {}
func (Kind_CharDevice) _Kind()  {}

func (Kind_File) String() string        { return "f" }
func (Kind_Dir) String() string         { return "d" }
func (Kind_Symlink) String() string     { return "l" }
func (Kind_Fifo) String() string        { return "p" }
func (Kind_Socket) String() string      { return "s" }
func (Kind_BlockDevice) String() string { return "b" }
func (Kind_CharDevice) String() string  { return "c" }

func (Kind_File) GoString() string        { return "file.Kind_File" }
func (Kind_Dir) GoString() string         { return "file.Kind_Dir" }
func (Kind_Symlink) GoString() string     { return "file.Kind_Symlink" }
func (Kind_Fifo) GoString() string        { return "file.Kind_Fifo" }
func (Kind_Socket) GoString() string      { return "file.Kind_Socket" }
func (Kind_BlockDevice) GoString() string { return "file.Kind_Device" }
func (Kind_CharDevice) GoString() string  { return "file.Kind_CharDevice" }

type File interface {
	io.Writer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

type Dir interface {
	// TODO: this doens't contain enough info for all kinds.  have extended options field?  use builder chain?  design decision.
	//
	// REVIEW: several (radically different) alternative ways we could go about this:
	// - Separate CreateFoo methods, one for each Kind
	// - Take the _entire_ Metadata struct as a param
	//   - ...this does *not* map onto a single atomic syscall on linux -- but is that what we want to orient this API around?  perhaps not.
	//   - ...this does *not* pleasantly address that different details are relevant per kind (linkname only sane for symlinks; dev numbers for devs, etc).
	// - Do builder chains
	// - Do some form of complex extended options field (functional options potentially too costly, but, something in that vein).
	// Remember: we *can* add "porcelain" layers to this as a functional API.  This can be resigned to being a plumbing interface, if that's appropriate and constructive (and leads to speed).
	Create(name Name, omode Openmode, kind Kind) (Handle, error)

	Open(name Name, omode Openmode) (Handle, error)

	Read() DirItr

	io.Closer

	// REVIEW: Does Dir (or Handle, for that matter) cache the Path used to create it?  Undetermined.
	//  Possibly useful, yes.  Otoh: makes efficient walks *much* harder (can't so easily loan out PathBuffer contents and expect it to be brief).
	//  Would rather not be forced to cache the Path.  Also makes some implications that aren't super truthy.  Undecided how else to handle yet, though.
	//   Some sort of design that does gives a reasonable way to have cursors over paths, and hold the handles in parallel, might be desirable.  (See similar remarks esp on Rename functions.)
}

type DirItr interface {
	Next() (Name, *Metadata, error) // REVIEW: metadata?  do you always have it?  This is a case where loading Metadata.Linkname immediately (at the cost of an extra syscall) might be undesirable.
	NextBrief() (Name, error)       // REVIEW: maybe these two variations of iteration solve the above problem.  Also saves flipping the whole Metadata struct, in case that's a concern.
	Done() bool

	// FUTURE: are there 'seek' methods we can support on this?  (I don't think there are for many filesystems, so if so, this might be an interface feature detection thing.)

	// Implementation note: your DirItr implementation should probably consider embedding a Metadata struct;
	//  it can then just keep returning a pointer to it during all responses to 'Next', and thereby never allocating.
}

type (
	Openmode interface {
		_Openmode()
	}

	// REVIEW: this is... going to be syntactically irritating to _users_ in a way that doing the kind enum wasn't.  People regularly bitmask these together and that's convenient.
	//   We need an equally convenient set-aggregating syntax, one way or another.  And ideally it's similarly cheap (zeroalloc).

	Openmode_Create   struct{}
	Openmode_Truncate struct{}
	Openmode_Readable struct{}
	Openmode_Writable struct{}
	// ... etc ...
)

type Metadata struct {
	Perms    Mode      // permission bits
	Uid      uint32    // user id of owner
	Gid      uint32    // group id of owner
	Size     int64     // length in bytes
	Linkname string    // if symlink: target of link
	Devmajor int64     // major number of character or block device
	Devminor int64     // minor number of character or block device
	Ctime    time.Time // time of "creation"
	Mtime    time.Time // time of modification
	Atime    time.Time // time of access

	// Note that some of the time fields come with major caveats on any practical system:
	//  - ctime -- generally cannot be *set*, only read.
	//  - atime -- does not necessary represent the last access
	//      (many filesystems have settings like "noatime" or "relatime"
	//        which minimize the setting of atime for performance reasons).
	//      Can be set (though it's often of dubious practical utility to do so).

	// Xattrs are *absent* from this structure.
	//  Xattrs demand significantly more effort to read on most real-world filesystems than the rest of this structure
	//   (e.g., on linux, a syscall *per attribute read* is required -- the cost scales linearly);
	//  therefore we do not engage in those costly operations without direct request,
	//  and that direct request goes through another interface than this.
}

type Xattrs interface {
	Supported() bool
	Enumerate() []string
	Lookup(k string) string
}

// Mode tracks the familiar posixy 0777 bitmask.
type Mode struct {
	bits uint16

	// Should this include sticky, setuid, setgid?
	//  Maybe so, because they're often serialized together.
	//  Maybe not, because this type is not meant for APIs, not for direct serialization.
	//  Maybe not, because we already know that setuid and setgid get Different treatment in many cases.
	//  Sticky might be subject to separate discussion than setuid and setgid.

	// Should this track a tidbit of metadata about whether it's "actually" 0755 or if the backing store only said "executable"?
	//  This feels like a neat idea.  We should try it.
	//   Likely to be troublesome?  Also plausible, yes.
}

func NewModeFromBits(bits uint16) Mode {
	return Mode{bits & 0777}
}

func (m Mode) Raw() uint16 {
	return m.bits
}
func (m Mode) String() string {
	var buf [9]byte
	const rwx = "rwxrwxrwx"
	for i, c := range rwx {
		if m.bits&(1<<uint(9-1-i)) != 0 {
			buf[i] = byte(c)
		} else {
			buf[i] = '-'
		}
	}
	return string(buf[:])
}
