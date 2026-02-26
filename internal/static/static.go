package static

import (
	"runtime"
	"errors"
)

const (
	APIVersion = "/api/v2" // we're on v2 now, v1 used SQLite for database backend.
	OldAPIVersion = "/api/v1" // this version was v1.x.x version of the FN server before the rewrite.
)

var (
	Version = "nightly-canary"
	GoVersion = runtime.Version()
	OS = runtime.GOOS
	OSArch = runtime.GOARCH
)

var (
	ErrNoCharacterVersions = errors.New("no character versions exist")
	ErrBadCharacterData = errors.New("malformed character data")
)