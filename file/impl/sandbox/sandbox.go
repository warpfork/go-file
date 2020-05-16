// sandbox provides implementations of the go-file interfaces
// that delegate to another implementation as the backing store,
// while adding constraints.
//
// Specifically, sandbox checks every path that's about to be traversed
// to see if it contains any symlinks, and if the symlinks would leave the
// cabinet root.  If it encounters these, it return an E_Breakout error.
// This can be useful if you want to semantically isolate a cabinet
// but are operating on osfs, and cannot use a chroot syscall to do it
// with kernel assistence.
//
// The checks done by sandbox are not safe in the face of concurrent
// modification of the backing filesystem.  (This is fairly unavoidable
// due to TOCTOU issues central to the design of many real filesystems.)
package sandbox
