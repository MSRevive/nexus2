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
  flagsDebug = auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile

  Log *auralog.Logger
)

func InitLogging(filename string, dir string) {
  file := &auralog.RotateWriter{
    Dir: dir,
    Filename: filename,
    ExTime: 24 * time.Hour,
    MaxSize: 5 * auralog.Megabyte,
  }

  Api = auralog.New(auralog.Config{
    Output: io.MultiWriter(os.Stdout, file),
    Prefix: "[API] ",
    Level: auralog.LogLevelDebug,
    Flag: flags,
    WarnFlag: flagsWarn,
    ErrorFlag: flagsError,
    DebugFlag: flagsDebug,
  })
}
