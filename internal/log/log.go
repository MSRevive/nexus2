package log

import (
	"os"
	"io"
	"time"

	"github.com/saintwish/auralog"
)

var (
	flags = auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds
	flagsWarn = auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds
	flagsError = auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile
	flagsDebug = auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile

	Log *auralog.Logger
)

func InitLogging(filename string, dir string, level string, expire string) {
	ex, _ := time.ParseDuration(expire)
	
	file := &auralog.RotateWriter{
		Dir: dir,
		Filename: filename,
		ExTime: ex,
		MaxSize: 5 * auralog.Megabyte,
	}

	Log = auralog.New(auralog.Config{
		Output: io.MultiWriter(os.Stdout, file),
		Prefix: "[API] ",
		Level: auralog.ToLogLevel(level),
		Flag: flags,
		WarnFlag: flagsWarn,
		ErrorFlag: flagsError,
		DebugFlag: flagsDebug,
	})
}
