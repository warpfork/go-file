package fs

import (
	"github.com/polydawn/go-file/file"
)

func WalkUsingItr(cab file.Cabinet, start file.Path) WalkItr {
	return WalkItr{}
}

type WalkItr struct{}

func Walk(cab file.Cabinet, start file.Path, cb func(fh file.Handle) error) error {
	panic("nyi")
}
