package main

import (
	"fmt"
	"time"
	"os"

	"github.com/msrevive/nexus2/cmd"
	"github.com/msrevive/nexus2/internal/system"
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
	fmt.Printf(spMsg, time.Now().Year(), system.Version, "\n\n")

	if err := cmd.Run(os.Args); err != nil {
		panic(err)
	}
}