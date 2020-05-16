package file

// All errors should have a `who string` field that ends up in the output as "%s reports:".
// This is useful so error messages can say "os reports: symlink target does not exist"
//  versus "sandboxing fs reports: symlink target does not exist"... both of which are useful, but very distinct.

// Platform-specific error codes that use magic numbers should print as "E_WHATEVER/123", or "UNKNOWN/123" if the magic number is not known.
