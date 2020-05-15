package file

import (
	"testing"

	. "github.com/warpfork/go-wish"
)

func TestModeStrings(t *testing.T) {
	Wish(t, NewModeFromBits(0777).String(), ShouldEqual, "rwxrwxrwx")
	Wish(t, NewModeFromBits(0755).String(), ShouldEqual, "rwxr-xr-x")
	Wish(t, NewModeFromBits(0644).String(), ShouldEqual, "rw-r--r--")
	Wish(t, NewModeFromBits(0664).String(), ShouldEqual, "rw-rw-r--")
	Wish(t, NewModeFromBits(0000).String(), ShouldEqual, "---------")
}
