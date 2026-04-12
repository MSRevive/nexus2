package main

import (
	"fmt"
	"time"
	"os"
	"runtime/debug"

	"github.com/msrevive/nexus2/cmd"
	"github.com/msrevive/nexus2/internal/static"
)

var spMsg string = `
    _   __                    ___
   / | / /__  _  ____  Nexus2|__ \
  /  |/ / _ \| |/_/ / / / ___/_/ /
 / /|  /  __/>  </ /_/ (__  ) __/
/_/ |_/\___/_/|_|\__,_/____/____/
Copyright © %d, Team MSRebirth
Version: %s
Website: https://msrebirth.net/
License: GPL-3.0 https://github.com/MSRevive/nexus2/blob/main/LICENSE %s
`

func main() {
	defer func() {
		if r := recover(); r != nil {
			crashLog(fmt.Errorf("panic: %v\n%s", r, debug.Stack()))
			panic(r) // re-panic to preserve default behavior
		}
	}()

	fmt.Printf(spMsg, time.Now().Year(), static.Version, "\n")

	if err := cmd.Run(os.Args); err != nil {
		crashLog(err)
		panic(err)
	}
}

func crashLog(err error) {
	path := fmt.Sprintf("crash-%d.log", time.Now().Unix())
	content := fmt.Sprintf("time: %s\nerror: %v\n", time.Now().Format(time.RFC3339), err)
	_ = os.WriteFile(path, []byte(content), 0644)
}