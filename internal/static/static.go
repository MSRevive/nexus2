package static

import (
	"runtime"
)

const (
	APIVersion = "/api/v2" // we're on v2 now, v1 used SQLite for database backend.
)

var (
	Version = "canary"
	GoVersion = runtime.Version()
	OS = runtime.GOOS
	OSArch = runtime.GOARCH
)