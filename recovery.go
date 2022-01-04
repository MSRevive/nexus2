package main

import (
  "os"
  "runtime"

  "github.com/msrevive/nexus2/log"
)

func panicRecovery() {
  if panic := recover(); panic != nil {
    log.Bot.Errorln("Nexus has encountered an unrecoverable error and as crashed.")
    log.Bot.Errorln("Crash Information: " + panic.(error).Error())

    stack := make([]byte, 65536)
    l := runtime.Stack(stack, true)

    log.Bot.Panic("Stack trace:\n" + string(stack[:l]))

    os.Exit(1)
  }
}
