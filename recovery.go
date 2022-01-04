package main

import (
  "os"
  "runtime"

  "github.com/msrevive/nexus2/log"
)

func panicRecovery() {
  if panic := recover(); panic != nil {
    log.Log.Errorln("Nexus has encountered an unrecoverable error and as crashed.")
    log.Log.Errorln("Crash Information: " + panic.(error).Error())

    stack := make([]byte, 65536)
    l := runtime.Stack(stack, true)

    log.Log.Panic("Stack trace:\n" + string(stack[:l]))

    os.Exit(1)
  }
}
